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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const adminRequired = true
const adminNotRequired = false
const anyRol = ""

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

type UserITP struct {
	Nombre          string `bson:"nombre"`
	Apellido        string `bson:"apellido"`
	Email           string `bson:"email"`
	EsAdministrador bool   `bson:"esAdministrador"`
	Rol             string `bson:"rol"`
	Token           string `bson:"token"`
}

type UserITPWithRecursosData struct {
	Nombre          string `bson:"nombre"`
	Apellido        string `bson:"apellido"`
	Email           string `bson:"email"`
	EsAdministrador bool   `bson:"esAdministrador"`
	Rol             string `bson:"rol"`
	IdEncripted     string
	IdSecuencial    int
}

type GoogleClaims struct {
	Email         string `bson:"email"`
	EmailVerified bool   `bson:"email_verified"`
	FirstName     string `bson:"given_name"`
	LastName      string `bson:"family_name"`
	jwt.StandardClaims
}

type Recursos struct {
	IdRecurso   int       `bson:"idRecurso"`
	Nombre      string    `bson:"nombre"`
	Apellido    string    `bson:"apellido"`
	Legajo      int       `bson:"legajo"`
	Mail        string    `bson:"mail"`
	Fecha       time.Time `bson:"date"`
	FechaString string    `bson:"fechaString"`
	Sueldo      int       `bson:"sueldo"`
	Rcc         []P       `bson:"p"`
}

type RecursosWithID struct {
	IdObject    primitive.ObjectID `bson:"_id"`
	IdRecurso   int                `bson:"idRecurso"`
	Nombre      string             `bson:"nombre"`
	Apellido    string             `bson:"apellido"`
	Legajo      int                `bson:"legajo"`
	Mail        string             `bson:"mail"`
	Fecha       time.Time          `bson:"date"`
	FechaString string             `bson:"fechaString"`
	Sueldo      int                `bson:"sueldo"`
	Rcc         []P                `bson:"p"`
}

type P struct {
	CcNum     string  `bson:"cc"`
	CcPorc    float32 `bson:"porcCC"`
	CcNombre  string  `bson:"ccNombre"`
	CcCliente string  `bson:"ccCliente"`
}

type Cecos struct {
	IdCecos     int    `bson:"idCecos"`
	Descripcion string `bson:"descripcioncecos"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Codigo      int    `bson:"codigo"`
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
		return err, email
	}
	//valida el mail
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	var usuario UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
	if err2 != nil {
		if os.Getenv("USE_RECURSOS_LIKE_USERS") == "1" {
			collRecurso := client.Database("portalDeNovedades").Collection("recursos")
			var recurso Recursos
			err2 = collRecurso.FindOne(context.TODO(), bson.M{"mail": email}).Decode(&recurso)
			if err2 != nil {
				return errors.New("email no encontrado"), ""
			}
			usuario.Email = recurso.Mail
			usuario.EsAdministrador = false
			usuario.Rol = ""
			_, err := coll.InsertOne(context.TODO(), usuario)
			if err != nil {
				return errors.New("error al ingresar usuario desde recursos"), ""
			}
		} else {
			return errors.New("email no encontrado"), ""
		}
	}

	if obligatorioAdministrador && usuario.EsAdministrador == false {
		return errors.New("el usuario no tiene permiso para esta acción, no es administrador"), "403"
	}

	if rolEsperado != "" {
		result := strings.Split(rolEsperado, ",")
		var encontradoElRol bool
		for _, rol := range result {
			if rol == usuario.Rol {
				encontradoElRol = true
			}
		}
		if !encontradoElRol {
			return errors.New("el usuario no tiene permiso para esta acción, no tiene el rol"), "403"
		}
	}

	return nil, email
}

func ValidacionDeUsuarioPropio(obligatorioAdministrador bool, rolEsperado string, token string) (error, string) {
	//valida el mail
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	var usuario UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"token": token}).Decode(&usuario)
	if err2 != nil {
		return errors.New("el usuario o token inexistente"), "401"
	}

	if obligatorioAdministrador && usuario.EsAdministrador == false {
		return errors.New("el usuario no tiene permiso para esta acción, no es administrador"), "403"
	}

	if rolEsperado != "" && rolEsperado == usuario.Rol {
		return errors.New("el usuario no tiene permiso para esta acción, no tiene el rol"), "403"
	}

	return nil, usuario.Email
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func Authorization(authHeader string, administrationRequired bool, rolRequired string) (error, int, bool) {
	if authHeader == "" {
		// el token no esta presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación"), fiber.StatusBadRequest, false
	}

	// parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	fmt.Println(tokenString)

	// valida el token
	err, codigo := ValidacionDeUsuarioPropio(administrationRequired, rolRequired, tokenString)
	if err != nil {
		if codigo != "" {
			codigoError, _ := strconv.Atoi(codigo)
			return err, codigoError, false
		}
		return err, fiber.StatusBadRequest, false
	}
	return nil, fiber.StatusAccepted, true
}

func InsertUserITP(c *fiber.Ctx) error {

	//obtiene los datos
	var userITP UserITP
	if err := c.BodyParser(&userITP); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	if userITP.Nombre == "" || userITP.Apellido == "" {
		coll := client.Database("portalDeNovedades").Collection("recursos")
		email := userITP.Email
		var recurso Recursos
		err2 := coll.FindOne(context.TODO(), bson.M{"mail": email}).Decode(&recurso)
		if err2 == nil {
			userITP.Nombre = recurso.Nombre
			userITP.Apellido = recurso.Apellido
		}
	}

	coll := client.Database("portalDeNovedades").Collection("usersITP")

	//inserta el usuario
	result, err := coll.InsertOne(context.TODO(), userITP)
	if err != nil {

		fmt.Print("fail")
		return c.SendString(err.Error())
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.SendString("ok")
}

func GetUserITP(c *fiber.Ctx) error {

	coll := client.Database("portalDeNovedades").Collection("usersITP")
	email := c.Params("email")
	var usuario UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
	if err2 != nil {
		return c.Status(200).SendString("usuario no encontrada")
	}
	return c.Status(200).JSON(usuario)
}

func GetSelfUserITP(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	//valida el token
	err, email := validacionDeUsuario(false, "", tokenString)
	if err != nil {
		if email != "" {
			codigoError, _ := strconv.Atoi(email)
			return c.Status(codigoError).SendString(err.Error())
		}
		return c.Status(404).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("usersITP")

	userITP := new(UserITP)
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&userITP)
	if err2 != nil {
		return c.Status(404).SendString("usuario no encontrado")
	}
	filter := bson.D{{"email", email}}
	update := bson.D{{"$set", bson.D{{"token", tokenString}}}}
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
	collR := client.Database("portalDeNovedades").Collection("recursos")
	recurso := new(RecursosWithID)
	err2 = collR.FindOne(context.TODO(), bson.M{"mail": email}).Decode(&recurso)

	userITPWithRecursosData := new(UserITPWithRecursosData)
	userITPWithRecursosData.Email = email
	userITPWithRecursosData.EsAdministrador = userITP.EsAdministrador
	userITPWithRecursosData.Rol = userITP.Rol
	idObjectHash, err := hashPassword(recurso.IdObject.Hex())
	if err != nil {
		return c.Status(401).SendString(err.Error())
	}
	userITPWithRecursosData.IdEncripted = idObjectHash
	userITPWithRecursosData.IdSecuencial = recurso.IdRecurso

	return c.Status(200).JSON(userITPWithRecursosData)
}

// obtener todos los usuarios
func GetUserITPAll(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	//valida el token
	err, codigo := validacionDeUsuario(true, "admin,servicios", tokenString)
	if err != nil {
		if codigo != "" {
			codigoError, _ := strconv.Atoi(codigo)
			return c.Status(codigoError).SendString(err.Error())
		}
		return c.Status(404).SendString(err.Error())
	}

	//obtener los usuarios
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	var usuarios []UserITP
	if err = cursor.All(context.Background(), &usuarios); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	return c.Status(200).JSON(usuarios)
}

func DeleteUserITP(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("usersITP")
	emailDelete := c.Params("email")
	result, err := coll.DeleteOne(context.TODO(), bson.M{"email": emailDelete})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("usuario eliminado")
}

func UpdateUserITP(c *fiber.Ctx) error {
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
	return c.Status(200).JSON(result)
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
	fmt.Print("Expiracion: ")
	fmt.Println(time.Unix(claims.ExpiresAt, 0))
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
		return errors.New("JWT is expired"), "401"
	}

	return nil, *&claims.Email
}
