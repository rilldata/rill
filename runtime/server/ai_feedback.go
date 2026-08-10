package server

import (
	"context"
	"errors"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// aiFeedbackDefaultPageSize is the default page size for ListAIFeedback.
const aiFeedbackDefaultPageSize = 100

func (s *Server) ListAIFeedback(ctx context.Context, req *runtimev1.ListAIFeedbackRequest) (*runtimev1.ListAIFeedbackResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ManageAIFeedback) {
		return nil, ErrForbidden
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	opts := &drivers.FindAIFeedbackOptions{
		Status: req.Status,
		Kind:   req.Kind,
		Limit:  pagination.ValidPageSize(req.PageSize, aiFeedbackDefaultPageSize),
	}
	if req.PageToken != "" {
		if err := pagination.UnmarshalPageToken(req.PageToken, &opts.AfterCreatedOn, &opts.AfterID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token: %s", err.Error())
		}
	}

	items, err := catalog.FindAIFeedback(ctx, opts)
	if err != nil {
		return nil, err
	}

	res := make([]*runtimev1.AIFeedback, len(items))
	for i, f := range items {
		res[i] = aiFeedbackToPB(f)
	}

	var nextPageToken string
	if len(items) == opts.Limit {
		last := items[len(items)-1]
		nextPageToken = pagination.MarshalPageToken(last.CreatedOn, last.ID)
	}

	return &runtimev1.ListAIFeedbackResponse{
		Feedback:      res,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) GetAIFeedback(ctx context.Context, req *runtimev1.GetAIFeedbackRequest) (*runtimev1.GetAIFeedbackResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ManageAIFeedback) {
		return nil, ErrForbidden
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	feedback, err := catalog.FindAIFeedbackByID(ctx, req.FeedbackId)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "feedback %q not found", req.FeedbackId)
		}
		return nil, err
	}

	res := &runtimev1.GetAIFeedbackResponse{
		Feedback: aiFeedbackToPB(feedback),
	}

	// Load the conversation transcript with reviewer access.
	// ReviewerAccess is safe here because we verified above that the session has a feedback item.
	// A missing session (e.g. expired after the feedback was resolved) is not an error; the response just has no transcript.
	session, err := s.ai.Session(ctx, &ai.SessionOptions{
		InstanceID:     req.InstanceId,
		SessionID:      feedback.SessionID,
		Claims:         claims,
		ReviewerAccess: true,
	})
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return res, nil
		}
		return nil, err
	}

	messages := session.Messages()
	messagePBs := make([]*runtimev1.Message, 0, len(messages))
	for _, msg := range messages {
		pb, err := messageToPB(session, msg)
		if err != nil {
			return nil, err
		}
		messagePBs = append(messagePBs, pb)
	}

	res.Conversation = sessionToPB(session.CatalogSession(), messagePBs)
	res.Messages = messagePBs
	return res, nil
}

func (s *Server) UpdateAIFeedbackStatus(ctx context.Context, req *runtimev1.UpdateAIFeedbackStatusRequest) (*runtimev1.UpdateAIFeedbackStatusResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ManageAIFeedback) {
		return nil, ErrForbidden
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	feedback, err := catalog.FindAIFeedbackByID(ctx, req.FeedbackId)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "feedback %q not found", req.FeedbackId)
		}
		return nil, err
	}

	// Idempotent: if the status is already the requested one, return the current item.
	// This makes concurrent resolution by multiple admins race benignly.
	if feedback.Status == req.Status {
		return &runtimev1.UpdateAIFeedbackStatusResponse{Feedback: aiFeedbackToPB(feedback)}, nil
	}

	feedback.Status = req.Status
	if req.Status == drivers.AIFeedbackStatusOpen {
		feedback.ResolvedBy = ""
		feedback.ResolvedOn = nil
	} else {
		now := time.Now()
		feedback.ResolvedBy = claims.UserID
		feedback.ResolvedOn = &now
	}

	if err := catalog.UpdateAIFeedback(ctx, feedback); err != nil {
		return nil, err
	}

	return &runtimev1.UpdateAIFeedbackStatusResponse{Feedback: aiFeedbackToPB(feedback)}, nil
}

func (s *Server) GenerateAIFeedbackFix(ctx context.Context, req *runtimev1.GenerateAIFeedbackFixRequest) (*runtimev1.GenerateAIFeedbackFixResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}
	if req.Feedback == nil {
		return nil, status.Error(codes.InvalidArgument, "missing feedback")
	}

	opts := &ai.SuggestFeedbackFixOptions{
		InstanceID:           req.InstanceId,
		Kind:                 req.Feedback.Kind,
		Sentiment:            req.Feedback.Sentiment,
		Categories:           req.Feedback.Categories,
		Comment:              req.Feedback.Comment,
		PredictedAttribution: req.Feedback.PredictedAttribution,
	}
	for _, msg := range req.Transcript {
		opts.Transcript = append(opts.Transcript, ai.SuggestFixTranscriptMessage{Role: msg.Role, Text: msg.Text})
	}

	res, err := ai.SuggestFeedbackFix(ctx, s.runtime, opts)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateAIFeedbackFixResponse{
		Action:             res.Action,
		Summary:            res.Summary,
		MetricsView:        res.MetricsView,
		FilePath:           res.FilePath,
		Instruction:        res.Instruction,
		MeasureYaml:        res.MeasureYAML,
		EvalQuestion:       res.EvalQuestion,
		EvalExpectedAnswer: res.EvalExpectedAnswer,
	}, nil
}

func (s *Server) GenerateAIEvalFix(ctx context.Context, req *runtimev1.GenerateAIEvalFixRequest) (*runtimev1.GenerateAIEvalFixResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	res, err := ai.SuggestEvalFix(ctx, s.runtime, &ai.SuggestEvalFixOptions{
		InstanceID: req.InstanceId,
		Eval:       req.Eval,
		Cases:      req.Cases,
	})
	if err != nil {
		return nil, err
	}

	resp := &runtimev1.GenerateAIEvalFixResponse{Analysis: res.Analysis}
	for _, p := range res.Proposals {
		resp.Proposals = append(resp.Proposals, &runtimev1.AIEvalFixProposal{
			FilePath:    p.FilePath,
			Reason:      p.Reason,
			Instruction: p.Instruction,
		})
	}
	return resp, nil
}

func aiFeedbackToPB(f *drivers.AIFeedback) *runtimev1.AIFeedback {
	res := &runtimev1.AIFeedback{
		Id:                   f.ID,
		ConversationId:       f.SessionID,
		TargetMessageId:      f.TargetMessageID,
		Kind:                 f.Kind,
		Sentiment:            f.Sentiment,
		Categories:           f.Categories,
		Comment:              f.Comment,
		PredictedAttribution: f.PredictedAttribution,
		SuggestedAction:      f.SuggestedAction,
		Status:               f.Status,
		OwnerId:              f.OwnerID,
		OwnerEmail:           f.OwnerEmail,
		ResolvedBy:           f.ResolvedBy,
		CreatedOn:            timestamppb.New(f.CreatedOn),
		UpdatedOn:            timestamppb.New(f.UpdatedOn),
		ConversationTitle:    f.SessionTitle,
	}
	if f.ResolvedOn != nil {
		res.ResolvedOn = timestamppb.New(*f.ResolvedOn)
	}
	return res
}
