package controllers

import (
	"encoding/json"
	"golang-poc/dao"
	"golang-poc/dto"
	u "golang-poc/utils"
	"net/http"
)

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	createAccountDto := &dto.CreateAccountDto{}
	err := json.NewDecoder(r.Body).Decode(createAccountDto)
	if err != nil {
		u.RespondWithError(w, u.Message(400, "Invalid request"), http.StatusBadRequest)
		return
	}

	if resp, ok := createAccountDto.Validate(); !ok {
		u.RespondWithError(w, resp, http.StatusBadRequest)
		return
	}

	resp, err := dao.Create(createAccountDto)

	if err != nil {
		u.RespondWithError(w, u.Message(400, err.Error()), http.StatusBadRequest)
	}

	u.Respond(w, resp)
}

var Authenticate = func(w http.ResponseWriter, r *http.Request) {

	account := &dao.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		u.RespondWithError(w, u.Message(400, "Invalid request"), http.StatusBadRequest)
		return
	}

	resp := dao.Login(account.Email, account.Password)
	u.RespondWithMessage(w, resp)
}
