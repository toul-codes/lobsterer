package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/spf13/viper"
)

// Cognito represents an abstraction of the Amazon Cognito identity provider service.
type Cognito struct {
	cip *cognitoidentityprovider.CognitoIdentityProvider
}

var userPoolID string
var clientID string
var jwksURL string
var keySet *jwk.Set

func init() {

	log.Info("Initializing Cognito")

	log.Info("Loading configuration")
	viper.SetConfigName("config") // config.toml
	viper.AddConfigPath(".")      // use working directory

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("error reading config file, %v", err)
		return
	}

	userPoolID = viper.GetString("cognito.userPoolID")
	clientID = viper.GetString("cognito.clientID")
	jwksURL = viper.GetString("cognito.jwksURL")

	log.Info("userPoolID: ", userPoolID)
	log.Info("clientID: ", clientID)
	log.Info("jwksURL: ", jwksURL)

	if err := loadKeySet(); err != nil {
		log.Error("Error: ", err)
	}
}

// loadKeySet caches the keyset so we don't have to make a request every time
// we want to verify a JWT
func loadKeySet() error {
	log.Info("Caching keyset")
	var err error
	keySet, err = jwk.FetchHTTP(jwksURL)
	if err != nil {
		return err
	}
	return nil
}

// NewCognito creates a new instance of the Cognito client
func NewCognito() *Cognito {

	c := &Cognito{}

	// Create Session
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	c.cip = cognitoidentityprovider.New(sess)

	return c
}

// SignUp creates a new Cognito user in the user pool, setting its status
// to CONFIRMED. Returns the authenticated user's JWT access token.
func (c *Cognito) SignUp(username string, password string, email string, fullName string) (string, error) {

	log.Info("AdminCreateUser: ", username)

	// Creates user in FORCE_CHANGE_PASSWORD state
	_, err := c.cip.AdminCreateUser(&cognitoidentityprovider.AdminCreateUserInput{
		Username:          aws.String(username),
		TemporaryPassword: aws.String(password),
		UserPoolId:        aws.String(userPoolID),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(fullName),
			},
		},
	})

	if err != nil {
		log.Error("Error: ", err.Error())
		return "", err
	}

	// Attempt login to get session value, which is used to confirm the user

	aia := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(username),
			"PASSWORD": aws.String(password),
		},
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	log.Info("AdminInitiateAuth: ", username)
	authresp, autherr := c.cip.AdminInitiateAuth(aia)

	log.Info("ChallengeName: ", aws.StringValue(authresp.ChallengeName))

	if autherr != nil {
		log.Warn(autherr.Error())
	}

	// Set user to CONFIRMED

	artaci := &cognitoidentityprovider.AdminRespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"), // Required
		ClientId:      aws.String(clientID),                // Required
		UserPoolId:    aws.String(userPoolID),              // Required
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(username),
			"NEW_PASSWORD": aws.String(password), // Required
		},
		Session: authresp.Session, // session value from AdminInitiateAuth
	}

	log.Info("AdminRespondToAuthChallenge: ", username)
	chalresp, err := c.cip.AdminRespondToAuthChallenge(artaci)

	if err != nil {
		log.Error(err.Error())
		return "", err.(awserr.Error).OrigErr()
	}

	idToken := aws.StringValue(chalresp.AuthenticationResult.IdToken)
	accessToken := aws.StringValue(chalresp.AuthenticationResult.AccessToken)

	log.Debug("ID Token: ", idToken)
	log.Debug("AccessToken: ", accessToken)

	return accessToken, nil
}

// SignIn authenticates a user and returns a JWT token
func (c *Cognito) SignIn(username string, password string) (string, error) {

	aia := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(username),
			"PASSWORD": aws.String(password),
		},
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	log.Info("AdminInitiateAuth: ", username)
	authresp, autherr := c.cip.AdminInitiateAuth(aia)

	if autherr != nil {
		log.Error(autherr.Error())
		return "", autherr
	}

	accessToken := aws.StringValue(authresp.AuthenticationResult.AccessToken)

	log.Debug("AccessToken: ", accessToken)

	return accessToken, nil
}

// ValidateToken validates a JWT token and returns the 'sub' claim.
func (c *Cognito) ValidateToken(jwtToken string) (string, error) {

	log.Debug("ValidateToken: ", jwtToken)

	token, err := jwt.Parse(jwtToken, c.getKey)
	if err != nil {
		return "", fmt.Errorf("Could not parse JWT: %v", err)
	}

	log.Debug("JWT signature: ", token.Signature)

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["token_use"] != "access" {
			return "", fmt.Errorf("token_use mismatch: %s", claims["token_use"])
		}

		return claims["sub"].(string), nil // Valid token
	}

	return "", nil // Invalid token
}

// getKey returns the key for validating in ValidateToken
func (c *Cognito) getKey(token *jwt.Token) (interface{}, error) {
	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have string kid")
	}

	log.Debug("kid: ", keyID)

	if key := keySet.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}

	return nil, errors.New("unable to find key")
}
