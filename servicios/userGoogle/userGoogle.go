package userGoogle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

type UserITP struct {
	Email           string `bson:"email"`
	EsAdministrador bool   `bson:"esAdministrador"`
	Rol             string `bson:"rol"`
}

type GoogleClaims struct {
	Email         string `bson:"email"`
	EmailVerified bool   `bson:"email_verified"`
	FirstName     string `bson:"given_name"`
	LastName      string `bson:"family_name"`
	jwt.StandardClaims
}

func getGooglePublicKey(keyID string) (string, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	myResp := map[string]string{}
	err = json.Unmarshal(dat, &myResp)
	if err != nil {
		return "", err
	}
	key, ok := myResp[keyID]
	if !ok {
		return "", errors.New("key not found")
	}
	return key, nil
}

func validacionDeUsuario(obligatorioAdministrador bool, rolEsperado string, token string) (error, string) {
	//valida el token
	err, email := ValidateGoogleJWT(token)
	if err != nil {
		return err, ""
	}
	//valida el mail
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	var usuario UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
	if err2 != nil {
		return errors.New("email no encontrado"), ""
	}

	if obligatorioAdministrador && usuario.EsAdministrador == false {
		return errors.New("el usuario no tiene permiso para esta acción, no es administrador"), ""
	}

	if rolEsperado != "" && rolEsperado == usuario.Rol {
		return errors.New("el usuario no tiene permiso para esta acción, no tiene el rol"), ""
	}

	return nil, email
}

func InsertUserITP(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	//valida el token
	err2, _ := validacionDeUsuario(true, "", tokenString)
	if err2 != nil {
		return c.Status(403).SendString(err2.Error())
	}

	//obtiene los datos
	var body UserITP
	userITP := new(UserITP)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	coll := client.Database("portalDeNovedades").Collection("usersITP")

	//inserta el usuario
	result, err := coll.InsertOne(context.TODO(), userITP)
	if err != nil {
		return c.SendString(err.Error())
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(userITP)
}

func GetUserITP(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	//valida el token
	err, _ := validacionDeUsuario(true, "", tokenString)
	if err != nil {
		return c.Status(403).SendString(err.Error())
	}

	coll := client.Database("portalDeNovedades").Collection("usersITP")
	email, _ := strconv.Atoi(c.Params("email"))
	var usuario UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
	if err2 != nil {
		return c.SendString("novedad no encontrada")
	}
	return c.JSON(usuario)
}

func GetSelfUserITP(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	//valida el token
	err, email := validacionDeUsuario(true, "", tokenString)
	if err != nil {
		return c.Status(403).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	userITP := new(UserITP)
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&userITP)
	if err2 != nil {
		return c.SendString("usuario no encontrado")
	}
	return c.JSON(userITP)
}

func DeleteUserITP(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	//valida el token
	err, _ := validacionDeUsuario(true, "", tokenString)

	coll := client.Database("portalDeNovedades").Collection("usersITP")
	emailDelete, _ := strconv.Atoi(c.Params("email"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"email": emailDelete})
	if err != nil {
		return c.SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("novedad eliminada")
}

func UpdateUserITP(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	//valida el token
	err, _ := validacionDeUsuario(true, "", tokenString)

	usuario := new(UserITP)
	if err := c.BodyParser(usuario); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("usersITP")

	fmt.Println(usuario)

	filter := bson.D{{"email", usuario.Email}}

	update := bson.D{{"$set", bson.D{{"esAdministrador", usuario.EsAdministrador}, {"rol", usuario.Rol}}}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	return c.JSON(result)
}

func ValidateGoogleJWT(tokenString string) (error, string) {
	claimsStruct := GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := getGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return err, ""
	}

	claims, ok := token.Claims.(*GoogleClaims)
	if !ok {
		return errors.New("Invalid Google JWT"), ""
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return errors.New("iss is invalid"), ""
	}

	if claims.Audience != os.Getenv("GOOGLEKEY") {
		return errors.New("aud is invalid"), ""
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return errors.New("JWT is expired"), ""
	}

	return nil, *&claims.Email
}

/*
func loginHandler(c *fiber.Ctx) {
	defer r.Body.Close()

	// parse the GoogleJWT that was POSTed from the front-end
	type parameters struct {
		GoogleJWT *string
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Couldn't decode parameters")
		return
	}

	// Validate the JWT is valid
	claims, err := auth.ValidateGoogleJWT(*params.GoogleJWT)
	if err != nil {
		respondWithError(w, 403, "Invalid google auth")
		return
	}
	if claims.Email != user.Email {
		respondWithError(w, 403, "Emails don't match")
		return
	}

	// create a JWT for OUR app and give it back to the client for future requests
	tokenString, err := auth.MakeJWT(claims.Email, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, 500, "Couldn't make authentication token")
		return
	}

	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	})
}
*/
