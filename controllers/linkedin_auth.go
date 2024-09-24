package controllers

import (
	"golang.org/x/oauth2"
	"os"
)

var linkedinOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	ClientID:     "LINKEDIN_ID",
	ClientSecret: "LINKEDIN_SECRET",
	Scopes:       []string{"r_liteprofile", "r_emailaddress"},
	Endpoint: oauth2.Endpoint{
		AuthURL:   "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL:  "https://www.linkedin.com/oauth/v2/accessToken",
		AuthStyle: oauth2.AuthStyleInParams,
	},
}
