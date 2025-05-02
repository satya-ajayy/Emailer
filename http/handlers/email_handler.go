package handlers

import (
	// Go Internal Packages
	"context"
	"net/http"

	// Local Packages
	errors "emailer/errors"
	models "emailer/models"
	utils "emailer/utils/helpers"
)

type EmailService interface {
	AddToKafka(ctx context.Context, mailQP models.MailQP) error
}

type EmailHandler struct {
	svc EmailService
}

func NewEmailHandler(svc EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) (response any, status int, err error) {
	var p models.MailQP
	if dErr := utils.GetSchemaDecoder().Decode(&p, r.URL.Query()); dErr != nil {
		return nil, http.StatusBadRequest, errors.InvalidParamsErr(dErr)
	}

	if err = p.Validate(); err != nil {
		return nil, http.StatusBadRequest, errors.ValidationFailedErr(err)
	}

	err = h.svc.AddToKafka(r.Context(), p)
	if err == nil {
		return map[string]interface{}{
			"message": "successfully added to kafka topic",
		}, http.StatusOK, nil
	}
	return
}
