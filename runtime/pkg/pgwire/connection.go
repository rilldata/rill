package pgwire

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/jackc/pgx/v5/pgproto3"
)

type connection struct {
	backend *pgproto3.Backend
	session Session
	ctx     context.Context
	cancel  *cancelEntry

	statements map[string]*preparedStatement
	portals    map[string]*portal
	txStatus   byte
	failed     bool
}

type preparedStatement struct {
	query         string
	parameterOIDs []uint32
	description   *Description
}

type portal struct {
	statement     *preparedStatement
	parameters    []Parameter
	resultFormats []int16
	rows          Rows
	cancel        context.CancelFunc
	rowCount      int64
	described     bool
}

func (c *connection) run() error {
	defer c.closePortals()
	for {
		message, err := c.backend.Receive()
		if err != nil {
			return err
		}

		if c.failed {
			switch message.(type) {
			case *pgproto3.Sync:
				c.failed = false
				c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
			case *pgproto3.Terminate:
				return nil
			case *pgproto3.Flush:
				if err := c.backend.Flush(); err != nil {
					return err
				}
			}
			continue
		}

		switch message := message.(type) {
		case *pgproto3.Query:
			if err := c.handleSimpleQuery(message.String); err != nil {
				return err
			}
		case *pgproto3.Parse:
			if err := c.handleParse(message); err != nil {
				c.extendedError(err)
			}
		case *pgproto3.Bind:
			if err := c.handleBind(message); err != nil {
				c.extendedError(err)
			}
		case *pgproto3.Describe:
			if err := c.handleDescribe(message); err != nil {
				c.extendedError(err)
			}
		case *pgproto3.Execute:
			if err := c.handleExecute(message); err != nil {
				c.extendedError(err)
			}
		case *pgproto3.Close:
			if err := c.handleClose(message); err != nil {
				c.extendedError(err)
			}
		case *pgproto3.Flush:
			if err := c.backend.Flush(); err != nil {
				return err
			}
		case *pgproto3.Sync:
			c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
			if err := c.backend.Flush(); err != nil {
				return err
			}
		case *pgproto3.Terminate:
			return nil
		default:
			c.extendedError(fmt.Errorf("unsupported frontend message %T", message))
		}
	}
}

func (c *connection) handleSimpleQuery(query string) error {
	c.closePortal("")
	delete(c.statements, "")

	query = strings.TrimSpace(query)
	if query == "" {
		c.backend.Send(&pgproto3.EmptyQueryResponse{})
		c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
		return c.backend.Flush()
	}

	if tag, ok := c.sessionCommand(query); ok {
		c.backend.Send(&pgproto3.CommandComplete{CommandTag: []byte(tag)})
		c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
		return c.backend.Flush()
	}

	queryCtx, cancel := context.WithCancel(c.ctx)
	c.cancel.set(cancel)
	defer func() {
		cancel()
		c.cancel.clear()
	}()
	rows, err := c.session.Query(queryCtx, query, nil, nil)
	if err != nil {
		sendError(c.backend, err, "XX000")
		c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
		return c.backend.Flush()
	}
	defer rows.Close()

	_, err = c.sendResult(rows, true, 0, nil)
	if err != nil {
		if isConnectionError(err) {
			return err
		}
		sendError(c.backend, err, "XX000")
	}
	c.backend.Send(&pgproto3.ReadyForQuery{TxStatus: c.txStatus})
	return c.backend.Flush()
}

func (c *connection) handleParse(message *pgproto3.Parse) error {
	if message.Name == "" {
		c.closePortal("")
	}
	statement := &preparedStatement{
		query:         message.Query,
		parameterOIDs: append([]uint32(nil), message.ParameterOIDs...),
	}
	if isSessionCommand(message.Query) {
		statement.description = &Description{ParameterOIDs: statement.parameterOIDs}
	} else if _, err := c.describe(statement); err != nil {
		return err
	}
	c.statements[message.Name] = statement
	c.backend.Send(&pgproto3.ParseComplete{})
	return nil
}

func (c *connection) handleBind(message *pgproto3.Bind) error {
	statement, ok := c.statements[message.PreparedStatement]
	if !ok {
		return protocolError("prepared statement %q does not exist", message.PreparedStatement)
	}
	if len(message.Parameters) != len(statement.parameterOIDs) {
		return protocolError("bind message supplies %d parameters, but prepared statement requires %d", len(message.Parameters), len(statement.parameterOIDs))
	}
	if len(message.ParameterFormatCodes) != 0 && len(message.ParameterFormatCodes) != 1 && len(message.ParameterFormatCodes) != len(message.Parameters) {
		return protocolError("bind message has %d parameter formats but %d parameters", len(message.ParameterFormatCodes), len(message.Parameters))
	}

	parameters := make([]Parameter, len(message.Parameters))
	for i, value := range message.Parameters {
		format := int16(0)
		if len(message.ParameterFormatCodes) == 1 {
			format = message.ParameterFormatCodes[0]
		} else if len(message.ParameterFormatCodes) > 1 {
			format = message.ParameterFormatCodes[i]
		}
		parameters[i] = Parameter{OID: statement.parameterOIDs[i], Format: format}
		if value != nil {
			parameters[i].Value = append([]byte(nil), value...)
		}
	}

	c.closePortal(message.DestinationPortal)
	c.portals[message.DestinationPortal] = &portal{
		statement:     statement,
		parameters:    parameters,
		resultFormats: append([]int16(nil), message.ResultFormatCodes...),
	}
	c.backend.Send(&pgproto3.BindComplete{})
	return nil
}

func (c *connection) handleDescribe(message *pgproto3.Describe) error {
	switch message.ObjectType {
	case 'S':
		statement, ok := c.statements[message.Name]
		if !ok {
			return protocolError("prepared statement %q does not exist", message.Name)
		}
		description, err := c.describe(statement)
		if err != nil {
			return err
		}
		c.backend.Send(&pgproto3.ParameterDescription{ParameterOIDs: description.ParameterOIDs})
		c.sendDescription(description.Fields)
	case 'P':
		portal, ok := c.portals[message.Name]
		if !ok {
			return protocolError("portal %q does not exist", message.Name)
		}
		description, err := c.describe(portal.statement)
		if err != nil {
			return err
		}
		fields, err := fieldsWithFormats(description.Fields, portal.resultFormats)
		if err != nil {
			return err
		}
		portal.described = true
		c.sendDescription(fields)
	default:
		return protocolError("invalid Describe object type %q", message.ObjectType)
	}
	return nil
}

func (c *connection) describe(statement *preparedStatement) (*Description, error) {
	if statement.description != nil {
		return statement.description, nil
	}
	queryCtx, cancel := context.WithCancel(c.ctx)
	c.cancel.set(cancel)
	defer func() {
		cancel()
		c.cancel.clear()
	}()
	description, err := c.session.Describe(queryCtx, statement.query, statement.parameterOIDs)
	if err != nil {
		return nil, err
	}
	if description == nil {
		description = &Description{ParameterOIDs: statement.parameterOIDs}
	}
	if len(description.ParameterOIDs) == 0 && len(statement.parameterOIDs) != 0 {
		description.ParameterOIDs = append([]uint32(nil), statement.parameterOIDs...)
	}
	statement.parameterOIDs = append([]uint32(nil), description.ParameterOIDs...)
	statement.description = description
	return description, nil
}

func (c *connection) sendDescription(fields []pgproto3.FieldDescription) {
	if len(fields) == 0 {
		c.backend.Send(&pgproto3.NoData{})
		return
	}
	c.backend.Send(&pgproto3.RowDescription{Fields: fields})
}

func (c *connection) handleExecute(message *pgproto3.Execute) error {
	portal, ok := c.portals[message.Portal]
	if !ok {
		return protocolError("portal %q does not exist", message.Portal)
	}

	if portal.rows == nil {
		if tag, ok := c.sessionCommand(portal.statement.query); ok {
			c.backend.Send(&pgproto3.CommandComplete{CommandTag: []byte(tag)})
			delete(c.portals, message.Portal)
			return nil
		}

		queryCtx, cancel := context.WithCancel(c.ctx)
		portal.cancel = cancel
		c.cancel.set(cancel)
		rows, err := c.session.Query(queryCtx, portal.statement.query, portal.parameters, portal.resultFormats)
		if err != nil {
			cancel()
			c.cancel.clear()
			return err
		}
		portal.rows = rows
	} else {
		c.cancel.set(portal.cancel)
	}
	defer c.cancel.clear()

	suspended, err := c.sendResult(portal.rows, false, message.MaxRows, portal)
	if err != nil {
		c.closePortal(message.Portal)
		return err
	}
	if !suspended {
		c.closePortal(message.Portal)
	}
	return nil
}

func (c *connection) sendResult(rows Rows, includeDescription bool, maxRows uint32, portal *portal) (bool, error) {
	if includeDescription {
		c.sendDescription(rows.Fields())
	}

	var sent uint32
	for maxRows == 0 || sent < maxRows {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return false, err
			}
			tag := rows.CommandTag()
			if tag == "" {
				count := int64(sent)
				if portal != nil {
					count += portal.rowCount
				}
				tag = fmt.Sprintf("SELECT %d", count)
			}
			c.backend.Send(&pgproto3.CommandComplete{CommandTag: []byte(tag)})
			return false, nil
		}
		c.backend.Send(&pgproto3.DataRow{Values: rows.Values()})
		sent++
	}

	if portal != nil {
		portal.rowCount += int64(sent)
	}
	c.backend.Send(&pgproto3.PortalSuspended{})
	return true, nil
}

func (c *connection) handleClose(message *pgproto3.Close) error {
	switch message.ObjectType {
	case 'S':
		if _, ok := c.statements[message.Name]; !ok {
			return protocolError("prepared statement %q does not exist", message.Name)
		}
		for name, portal := range c.portals {
			if portal.statement == c.statements[message.Name] {
				c.closePortal(name)
			}
		}
		delete(c.statements, message.Name)
	case 'P':
		if _, ok := c.portals[message.Name]; !ok {
			return protocolError("portal %q does not exist", message.Name)
		}
		c.closePortal(message.Name)
	default:
		return protocolError("invalid Close object type %q", message.ObjectType)
	}
	c.backend.Send(&pgproto3.CloseComplete{})
	return nil
}

func (c *connection) sessionCommand(query string) (string, bool) {
	trimmed := strings.TrimSpace(strings.TrimSuffix(query, ";"))
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", false
	}
	command := strings.ToUpper(fields[0])
	switch command {
	case "BEGIN", "START":
		c.txStatus = 'T'
		return "BEGIN", true
	case "COMMIT", "END":
		c.txStatus = 'I'
		return "COMMIT", true
	case "ROLLBACK":
		c.txStatus = 'I'
		return "ROLLBACK", true
	case "SET", "RESET", "DISCARD":
		return command, true
	default:
		return "", false
	}
}

func isSessionCommand(query string) bool {
	fields := strings.Fields(strings.TrimSpace(strings.TrimSuffix(query, ";")))
	if len(fields) == 0 {
		return false
	}
	switch strings.ToUpper(fields[0]) {
	case "BEGIN", "START", "COMMIT", "END", "ROLLBACK", "SET", "RESET", "DISCARD":
		return true
	default:
		return false
	}
}

func (c *connection) extendedError(err error) {
	sendError(c.backend, err, "XX000")
	c.failed = true
}

func (c *connection) closePortal(name string) {
	portal, ok := c.portals[name]
	if !ok {
		return
	}
	if portal.rows != nil {
		_ = portal.rows.Close()
	}
	if portal.cancel != nil {
		portal.cancel()
	}
	delete(c.portals, name)
}

func (c *connection) closePortals() {
	for name := range c.portals {
		c.closePortal(name)
	}
}

func fieldsWithFormats(fields []pgproto3.FieldDescription, formats []int16) ([]pgproto3.FieldDescription, error) {
	result := append([]pgproto3.FieldDescription(nil), fields...)
	switch len(formats) {
	case 0:
		for i := range result {
			result[i].Format = 0
		}
	case 1:
		if formats[0] != 0 && formats[0] != 1 {
			return nil, protocolError("unknown result format code %d", formats[0])
		}
		for i := range result {
			result[i].Format = formats[0]
		}
	default:
		if len(formats) != len(result) {
			return nil, protocolError("bind message has %d result formats but query has %d columns", len(formats), len(result))
		}
		for i, format := range formats {
			if format != 0 && format != 1 {
				return nil, protocolError("unknown result format code %d", format)
			}
			result[i].Format = format
		}
	}
	return result, nil
}

func sendError(backend *pgproto3.Backend, err error, defaultCode string) {
	code := defaultCode
	var pgErr *Error
	if errors.As(err, &pgErr) && pgErr.Code != "" {
		code = pgErr.Code
	}
	backend.Send(&pgproto3.ErrorResponse{
		Severity:            "ERROR",
		SeverityUnlocalized: "ERROR",
		Code:                code,
		Message:             err.Error(),
	})
}

// Error is an error with a PostgreSQL SQLSTATE code.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func protocolError(format string, args ...any) error {
	return &Error{Code: "08P01", Message: fmt.Sprintf(format, args...)}
}

func isConnectionError(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled)
}
