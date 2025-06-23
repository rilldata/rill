package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/encoding/protojson"
)

type catalogStore struct {
	*connection
	instanceID string
}

func (c *catalogStore) NextControllerVersion(ctx context.Context) (int64, error) {
	_, err := c.db.ExecContext(ctx, "INSERT OR IGNORE INTO controller_version(instance_id, version) VALUES (?, 0)", c.instanceID)
	if err != nil {
		return 0, err
	}

	_, err = c.db.ExecContext(ctx, "UPDATE controller_version SET version = version + 1 WHERE instance_id=?", c.instanceID)
	if err != nil {
		return 0, err
	}

	// TODO: Get it transactionally
	var version int64
	err = c.db.QueryRowContext(ctx, "SELECT version FROM controller_version WHERE instance_id=?", c.instanceID).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (c *catalogStore) CheckControllerVersion(ctx context.Context, v int64) error {
	var version int64
	err := c.db.QueryRowContext(ctx, "SELECT version FROM controller_version WHERE instance_id=?", c.instanceID).Scan(&version)
	if err != nil {
		return err
	}

	if version != v {
		return drivers.ErrInconsistentControllerVersion
	}

	return nil
}

func (c *catalogStore) FindResources(ctx context.Context) ([]drivers.Resource, error) {
	rows, err := c.db.QueryxContext(ctx, "SELECT kind, name, data FROM catalogv2 WHERE instance_id=? ORDER BY kind, lower(name)", c.instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.Resource
	for rows.Next() {
		r := drivers.Resource{}
		err := rows.Scan(&r.Kind, &r.Name, &r.Data)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogStore) CreateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO catalogv2(instance_id, kind, name, data, created_on, updated_on) VALUES (?, ?, ?, ?, ?, ?)",
		c.instanceID,
		r.Kind,
		r.Name,
		r.Data,
		now,
		now,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(
		ctx,
		"UPDATE catalogv2 SET name=?, data=?, updated_on=? WHERE instance_id=? AND kind=? AND lower(name)=lower(?)",
		r.Name,
		r.Data,
		time.Now(),
		c.instanceID,
		r.Kind,
		r.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteResource(ctx context.Context, v int64, k, n string) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, "DELETE FROM catalogv2 WHERE instance_id=? AND kind=? AND lower(name)=lower(?)", c.instanceID, k, n)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteResources(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM catalogv2 WHERE instance_id=?", c.instanceID)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) FindModelPartitions(ctx context.Context, opts *drivers.FindModelPartitionsOptions) ([]drivers.ModelPartition, error) {
	var qry strings.Builder
	var args []any

	qry.WriteString("SELECT key, data_json, idx, watermark, executed_on, error, elapsed_ms FROM model_partitions WHERE instance_id=? AND model_id=?")
	args = append(args, c.instanceID, opts.ModelID)

	if opts.WhereErrored {
		qry.WriteString(" AND error != ''")
	}

	if opts.WherePending {
		qry.WriteString(" AND executed_on IS NULL")
	}

	if opts.AfterIndex != 0 || opts.AfterKey != "" {
		qry.WriteString(" AND (idx > ? OR (idx = ? AND key > ?))")
		args = append(args, opts.AfterIndex, opts.AfterIndex, opts.AfterKey)
	}

	qry.WriteString(" ORDER BY idx, key")

	if opts.Limit != 0 {
		qry.WriteString(" LIMIT ?")
		args = append(args, opts.Limit)
	}

	rows, err := c.db.QueryContext(ctx, qry.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.ModelPartition
	for rows.Next() {
		var elapsedMs int64
		r := drivers.ModelPartition{}
		err := rows.Scan(&r.Key, &r.DataJSON, &r.Index, &r.Watermark, &r.ExecutedOn, &r.Error, &elapsedMs)
		if err != nil {
			return nil, err
		}
		r.Elapsed = time.Duration(elapsedMs) * time.Millisecond
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogStore) FindModelPartitionsByKeys(ctx context.Context, modelID string, keys []string) ([]drivers.ModelPartition, error) {
	// We can't pass a []string as a bound parameter, so we have to build a query with a corresponding number of placeholders.
	var qry strings.Builder
	var args []any
	qry.WriteString("SELECT key, data_json, idx, watermark, executed_on, error, elapsed_ms FROM model_partitions WHERE instance_id=? AND model_id=? AND key IN (")
	args = append(args, c.instanceID, modelID)

	qry.Grow(len(keys)*2 + 14) // Makes room for one ",?" per key plus the ORDER BY clause
	for i, k := range keys {
		if i == 0 {
			qry.WriteString("?")
		} else {
			qry.WriteString(",?")
		}
		args = append(args, k)
	}
	qry.WriteString(") ORDER BY key")

	rows, err := c.db.QueryxContext(ctx, qry.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.ModelPartition
	for rows.Next() {
		var elapsedMs int64
		r := drivers.ModelPartition{}
		err := rows.Scan(&r.Key, &r.DataJSON, &r.Index, &r.Watermark, &r.ExecutedOn, &r.Error, &elapsedMs)
		if err != nil {
			return nil, err
		}
		r.Elapsed = time.Duration(elapsedMs) * time.Millisecond
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogStore) CheckModelPartitionsHaveErrors(ctx context.Context, modelID string) (bool, error) {
	rows, err := c.db.QueryContext(
		ctx,
		"SELECT 1 FROM model_partitions WHERE instance_id=? AND model_id=? AND error != '' LIMIT 1",
		c.instanceID,
		modelID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var hasErrors bool
	if rows.Next() {
		hasErrors = true
	}

	if rows.Err() != nil {
		return false, err
	}

	return hasErrors, nil
}

func (c *catalogStore) InsertModelPartition(ctx context.Context, modelID string, partition drivers.ModelPartition) error {
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO model_partitions(instance_id, model_id, key, data_json, idx, watermark, executed_on, error, elapsed_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.instanceID,
		modelID,
		partition.Key,
		partition.DataJSON,
		partition.Index,
		partition.Watermark,
		partition.ExecutedOn,
		partition.Error,
		partition.Elapsed.Milliseconds(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartition(ctx context.Context, modelID string, partition drivers.ModelPartition) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET data_json=?, idx=?, watermark=?, executed_on=?, error=?, elapsed_ms=? WHERE instance_id=? AND model_id=? AND key=?",
		partition.DataJSON,
		partition.Index,
		partition.Watermark,
		partition.ExecutedOn,
		partition.Error,
		partition.Elapsed.Milliseconds(),
		c.instanceID,
		modelID,
		partition.Key,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartitionPending(ctx context.Context, modelID, partitionKey string) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET executed_on=NULL WHERE instance_id=? AND model_id=? AND key=?",
		c.instanceID,
		modelID,
		partitionKey,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartitionsPendingIfError(ctx context.Context, modelID string) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET executed_on=NULL WHERE instance_id=? AND model_id=? AND error != ''",
		c.instanceID,
		modelID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteModelPartitions(ctx context.Context, modelID string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM model_partitions WHERE instance_id=? AND model_id=?", c.instanceID, modelID)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) FindInstanceHealth(ctx context.Context, instanceID string) (*drivers.InstanceHealth, error) {
	var h drivers.InstanceHealth
	err := c.db.QueryRowContext(ctx, "SELECT health_json, updated_on FROM instance_health WHERE instance_id=?", instanceID).Scan(&h.HealthJSON, &h.UpdatedOn)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (c *catalogStore) UpsertInstanceHealth(ctx context.Context, h *drivers.InstanceHealth) error {
	_, err := c.db.ExecContext(ctx, `INSERT INTO instance_health(instance_id, health_json, updated_on) Values (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(instance_id) DO UPDATE SET health_json=excluded.health_json, updated_on=excluded.updated_on;
	`, h.InstanceID, h.HealthJSON)
	return err
}

// ListConversations fetches all conversations in an instance for a given owner.
func (c *catalogStore) ListConversations(ctx context.Context, ownerID string) ([]*runtimev1.Conversation, error) {
	rows, err := c.db.QueryContext(ctx, `
        SELECT conversation_id, owner_id, title, created_on, updated_on
        FROM conversations
        WHERE instance_id = ? AND owner_id = ?
        ORDER BY updated_on DESC
    `, c.instanceID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*runtimev1.Conversation
	for rows.Next() {
		var conv runtimev1.Conversation
		var createdOn, updatedOn time.Time
		if err := rows.Scan(&conv.Id, &conv.OwnerId, &conv.Title, &createdOn, &updatedOn); err != nil {
			return nil, err
		}

		conv.CreatedOn = createdOn.Format(time.RFC3339)
		conv.UpdatedOn = updatedOn.Format(time.RFC3339)
		result = append(result, &conv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ensure we return an empty slice instead of nil
	if result == nil {
		result = []*runtimev1.Conversation{}
	}

	return result, nil
}

// GetConversation fetches a conversation by ID.
func (c *catalogStore) GetConversation(ctx context.Context, conversationID string) (*runtimev1.Conversation, error) {
	row := c.db.QueryRowContext(ctx, `
        SELECT conversation_id, owner_id, title, created_on, updated_on
        FROM conversations
        WHERE instance_id = ? AND conversation_id = ?
    `, c.instanceID, conversationID)
	var conv runtimev1.Conversation
	var createdOn, updatedOn time.Time
	if err := row.Scan(&conv.Id, &conv.OwnerId, &conv.Title, &createdOn, &updatedOn); err != nil {
		return nil, err
	}

	conv.CreatedOn = createdOn.Format(time.RFC3339)
	conv.UpdatedOn = updatedOn.Format(time.RFC3339)

	// Fetch messages for this conversation
	messages, err := c.ListMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	conv.Messages = messages

	return &conv, nil
}

// CreateConversation inserts a new conversation.
func (c *catalogStore) CreateConversation(ctx context.Context, ownerID, title string) (string, error) {
	conversationID := uuid.NewString()
	_, err := c.db.ExecContext(ctx, `
        INSERT INTO conversations (instance_id, conversation_id, owner_id, title, created_on, updated_on)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `, c.instanceID, conversationID, ownerID, title)
	return conversationID, err
}

// ListMessages fetches all messages for a conversation, ordered by sequence number.
func (c *catalogStore) ListMessages(ctx context.Context, conversationID string) ([]*runtimev1.Message, error) {
	rows, err := c.db.QueryContext(ctx, `
        SELECT message_id, role, content_json, created_on, updated_on, seq_num
        FROM messages
        WHERE instance_id = ? AND conversation_id = ?
        ORDER BY seq_num ASC
    `, c.instanceID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*runtimev1.Message
	for rows.Next() {
		var msg runtimev1.Message
		var contentJSONStr string
		var createdOn, updatedOn time.Time
		var seqNum int
		if err := rows.Scan(&msg.Id, &msg.Role, &contentJSONStr, &createdOn, &updatedOn, &seqNum); err != nil {
			return nil, err
		}

		// Parse content JSON into ContentBlock array using protojson
		if contentJSONStr != "" {
			// Create a temporary message with the content array wrapped
			wrappedJSON := fmt.Sprintf(`{"content": %s}`, contentJSONStr)
			tempMsg := &runtimev1.Message{}
			if err := protojson.Unmarshal([]byte(wrappedJSON), tempMsg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal content JSON: %w", err)
			}
			msg.Content = tempMsg.Content
		}

		msg.CreatedOn = createdOn.Format(time.RFC3339)
		msg.UpdatedOn = updatedOn.Format(time.RFC3339)
		result = append(result, &msg)
	}
	return result, nil
}

// AddMessage inserts a new message into a conversation.
func (c *catalogStore) AddMessage(ctx context.Context, conversationID, role string, content []*runtimev1.ContentBlock, parentMessageID *string) (string, error) {
	messageID := uuid.NewString()

	// Serialize content to JSON using protojson
	tempMsg := &runtimev1.Message{Content: content}
	contentJSON, err := protojson.Marshal(tempMsg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	// Extract just the content field from the JSON
	var tempObj map[string]interface{}
	if err := json.Unmarshal(contentJSON, &tempObj); err != nil {
		return "", fmt.Errorf("failed to parse temp JSON: %w", err)
	}

	contentOnlyJSON, err := json.Marshal(tempObj["content"])
	if err != nil {
		return "", fmt.Errorf("failed to marshal content only: %w", err)
	}

	// Auto-calculate seq_num using a subquery - this is atomic and race-condition safe
	_, err = c.db.ExecContext(ctx, `
        INSERT INTO messages (instance_id, conversation_id, seq_num, message_id, role, content_json, created_on, updated_on)
        VALUES (?, ?, 
            (SELECT COALESCE(MAX(seq_num), 0) + 1 FROM messages WHERE instance_id = ? AND conversation_id = ?),
            ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `, c.instanceID, conversationID, c.instanceID, conversationID, messageID, role, string(contentOnlyJSON))
	return messageID, err
}
