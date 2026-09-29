// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routers

import (
	stdcontext "context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func fastAutoSignin(ctx *context.Context) (string, error) {
	userId := getSessionUser(ctx)
	if userId == "" || isSessionExpired(ctx) {
		return "", nil
	}

	clientId := ctx.Input.Query("client_id")
	responseType := ctx.Input.Query("response_type")
	redirectUri := ctx.Input.Query("redirect_uri")
	scope := ctx.Input.Query("scope")
	state := ctx.Input.Query("state")
	nonce := ctx.Input.Query("nonce")
	codeChallenge := ctx.Input.Query("code_challenge")
	resource := ctx.Input.Query("resource")
	if clientId == "" || responseType != "code" || redirectUri == "" {
		return "", nil
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", nil
	}

	if !application.EnableAutoSignin {
		return "", nil
	}

	user, err := object.GetUser(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}

	isAllowed, err := isFastAutoSigninAllowed(ctx, user, application)
	if err != nil {
		return "", err
	}
	if !isAllowed {
		return "", nil
	}

	consentRequired, err := object.CheckConsentRequired(user, application, scope)
	if err != nil {
		return "", err
	}

	if consentRequired {
		return "", nil
	}

	code, err := object.GetOAuthCode(userId, clientId, "", "autoSignin", responseType, redirectUri, scope, state, nonce, codeChallenge, resource, ctx.Input.CruSession.SessionID(stdcontext.Background()), ctx.Request.Host, getAcceptLanguage(ctx))
	if err != nil {
		return "", err
	} else if code.Message != "" {
		return "", errors.New(code.Message)
	}

	sep := "?"
	if strings.Contains(redirectUri, "?") {
		sep = "&"
	}
	res := fmt.Sprintf("%s%scode=%s&state=%s", redirectUri, sep, url.QueryEscape(code.Code), url.QueryEscape(state))
	return res, nil
}

func isFastAutoSigninAllowed(ctx *context.Context, user *object.User, application *object.Application) (bool, error) {
	if user.NeedUpdatePassword {
		return false, nil
	}

	err := object.CheckApplicationSignin(application, user, util.GetClientIpFromRequest(ctx.Request), getAcceptLanguage(ctx))
	if err != nil {
		return false, nil
	}

	organization, err := object.GetOrganizationByUser(user)
	if err != nil {
		return false, err
	}
	if object.IsNeedPromptMfa(organization, user) {
		return false, nil
	}

	if user.Type == "paid-user" && !user.IsGlobalAdmin() && !user.IsAdmin {
		return hasActiveSubscription(user)
	}
	return true, nil
}

func hasActiveSubscription(user *object.User) (bool, error) {
	subscriptions, err := object.GetSubscriptionsByUser(user.Owner, user.Name)
	if err != nil {
		return false, err
	}

	for _, subscription := range subscriptions {
		if subscription.State == object.SubStateActive {
			return true, nil
		}
	}
	return false, nil
}

func isSessionExpired(ctx *context.Context) bool {
	session, ok := ctx.Input.Session("SessionData").(string)
	if !ok {
		return false
	}

	sessionData := &controllers.SessionData{}
	if err := util.JsonToStruct(session, sessionData); err != nil {
		return true
	}
	return sessionData.ExpireTime != 0 && sessionData.ExpireTime < time.Now().Unix()
}

func StaticFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path

	if urlPath == "/.well-known/acme-challenge/filename" {
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "acme-challenge", time.Now(), strings.NewReader("content"))
		return
	}

	if strings.HasPrefix(urlPath, "/api/") || strings.HasPrefix(urlPath, "/.well-known/") {
		return
	}
	if strings.HasPrefix(urlPath, "/files/") {
		return
	}
	serveHeadlessResponse(ctx)
}

func serveHeadlessResponse(ctx *context.Context) {
	ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	ctx.ResponseWriter.WriteHeader(http.StatusNotImplemented)
	_, _ = ctx.ResponseWriter.Write([]byte(`{"error":"headless_ui_required","message":"Casdoor core is running without a bundled UI; configure an external authentication UI for browser flows."}`))
}
