package controllers

import "golang.org/x/oauth2"

var linkedinOauthConfig = &oauth2.Config{
	ClientID:     "<client_id>",
	ClientSecret: "<client_secret",
	Scopes:       []string{"r_liteprofile", "r_emailaddress"},
	Endpoint: oauth2.Endpoint{
		AuthURL:   "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL:  "https://www.linkedin.com/oauth/v2/accessToken",
		AuthStyle: oauth2.AuthStyleInParams,
	},
	RedirectURL: "https://localhost:8080/auth/linkedin/callback",
}
