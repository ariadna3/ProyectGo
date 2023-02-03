package user

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type User struct {
	UserId   int    `json:"userId" gorm:"primaryKey"`
	Email    string `json:"email"`
	User     string `json:"user"`
	Nombre   string `json:"nombre"`
	Password string `json:"password"`
	Apellido string `json:"apellido"`
}

type Token_auth struct {
	TokenId    int    `json:"token_id" gorm:"primaryKey"`
	Token      string `json:"token"`
	UserId     int    `json:"user_id"`
	Generacion string `json:"generacion"`
	Expiracion string `json:"expiracion"`
}

type Cecos struct {
	IdCeco      int    `bson:"id_ceco"`
	Descripcion string `bson:"descripcion"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Cuit        int    `bson:"cuit"`
}

type Distribuciones struct {
	Porcentaje float64 `bson:"porcentaje"`
	Ceco       Cecos   `bson:"ceco"`
}

type Novedades struct {
	IdSecuencial          int              `bson:"idSecuencial"`
	Tipo                  string           `bson:"tipo"`
	Fecha                 string           `bson:"fecha"`
	Hora                  string           `bson:"hora"`
	Usuario               string           `bson:"usuario"`
	Proveedor             string           `bson:"proveedor"`
	Periodo               string           `bson:"periodo"`
	ImporteTotal          float64          `bson:"importeTotal"`
	ConceptoDeFacturacion string           `bson:"conceptoDeFacturacion"`
	Adjuntos              []string         `bson:"adjuntos"`
	Distribuciones        []Distribuciones `bson:"distribuciones"`
	Comentarios           string           `bson:"comentarios"`
	Promovido             bool             `bson:"promovido"`
	Cliente               string           `bson:"cliente"`
	Estado                string           `bson:"estado"`
	Motivo                string           `bson:"motivo"`
	EnviarA               string           `bson:"enviarA"`
	Contacto              string           `bson:"contacto"`
}

const (
	Pendiente = "pendiente"
	Aceptada  = "aceptada"
	Rechazada = "rechazada"
)

type TipoNovedad struct {
	Tipo        string `bson:"tipo"`
	Descripcion string `bson:"descripcion"`
}

type Actividades struct {
	IdNovedad int    `bson:"idNovedad"`
	Usuario   string `bson:"usuario"`
	Fecha     string `bson:"fecha"`
	Hora      string `bson:"hora"`
	Actividad string `bson:"actividad"`
}

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	NumeroDoc   int    `bson:"numeroDoc"`
	RazonSocial string `bson:"razonSocial"`
}

var store *session.Store = session.New()
var dbUser *gorm.DB
var client *mongo.Client

var maxAge int32 = 86400 * 30 // 30 days
var isProd bool = false       // Set to true when serving over https

func ConnectMariaDb(db *gorm.DB) {
	dbUser = db
	dbUser.AutoMigrate(&User{})
	dbUser.AutoMigrate(&Token_auth{})
}

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

func ShowGoogleAuthentication(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Inicializar",
	})
}

func UpdateEstadoNovedades(c *fiber.Ctx) error {
	//se obtiene el id
	idNumber, _ := strconv.Atoi(c.Params("id"))
	//se obtiene el estado
	estado := c.Params("estado")
	//se conecta a la DB
	coll := client.Database("portalDeNovedades").Collection("novedades")

	//Verifica que el estado sea uno valido
	if estado != Pendiente && estado != Aceptada && estado != Rechazada {
		return c.SendString("estado no valido")
	}

	//crea el filtro
	filter := bson.D{{"idSecuencial", idNumber}}
	update := bson.D{{"$set", bson.D{{"estado", estado}}}}

	//le dice que es lo que hay que modificar y con que
	update = bson.D{{"$set", bson.D{{"motivo", novedad.Estado}}}}

	//hace la modificacion
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	//devuelve el resultado
	return c.JSON(result)
}

func UpdateMotivoNovedades(c *fiber.Ctx) error {
	//se obtiene el id
	idNumber, _ := strconv.Atoi(c.Params("id"))
	//se obtiene el motivo
	motivo := c.Params("motivo")
	//se conecta a la DB
	coll := client.Database("portalDeNovedades").Collection("novedades")

	//crea el filtro
	filter := bson.D{{"idSecuencial", idNumber}}
	update := bson.D{{"$set", bson.D{{"motivo", motivo}}}}

	//le dice que es lo que hay que modificar y con que
	update = bson.D{{"$set", bson.D{{"motivo", novedad.Motivo}}}}

	//hace la modificacion
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	//devuelve el resultado
	return c.JSON(result)
}

// Busqueda con parametros Novedades
func GetGreddy(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	var busqueda bson.M
	if c.Query("idSecuencial") != "" {
		busqueda["idSecuencial"], _ = strconv.Atoi(c.Query("idSecuencial"))
	}

	if c.Query("tipo") != "" {
		busqueda["tipo"] = c.Query("tipo")
	}
	if c.Query("fecha") != "" {
		busqueda["fecha"] = c.Query("fecha")
	}
	if c.Query("hora") != "" {
		busqueda["hora"] = c.Query("hora")
	}
	if c.Query("usuario") != "" {
		busqueda["usuario"] = c.Query("usuario")
	}
	if c.Query("proveedor") != "" {
		busqueda["proveedor"] = c.Query("proveedor")
	}
	if c.Query("periodo") != "" {
		busqueda["periodo"] = c.Query("periodo")
	}
	if c.Query("conceptoDeFacturacion") != "" {
		busqueda["conceptoDeFacturacion"] = c.Query("conceptoDeFacturacion")
	}
	if c.Query("comentarios") != "" {
		busqueda["comentarios"] = c.Query("comentarios")
	}
	if c.Query("cliente") != "" {
		busqueda["cliente"] = c.Query("cliente")
	}
	fmt.Println(coll)
	return c.JSON(busqueda)
}

// Novedades
// insertar novedad
func InsertNovedad(c *fiber.Ctx) error {
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	if novedad.Estado != Pendiente && novedad.Estado != Aceptada && novedad.Estado != Rechazada {
		novedad.Estado = Pendiente
	}

	coll := client.Database("portalDeNovedades").Collection("novedades")

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idSecuencial", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Novedades
	cursor.All(context.TODO(), &results)

	novedad.IdSecuencial = results[0].IdSecuencial + 1
	result, err := coll.InsertOne(context.TODO(), novedad)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(novedad)
}

// obtener novedad por id
func GetNovedades(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	return c.JSON(novedades)
}

// obtener novedad por tipo
func GetTipoNovedad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("tipoNovedad")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	fmt.Println("tipos")
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var tipoNovedad []TipoNovedad
	fmt.Println(tipoNovedad)

	if err = cursor.All(context.Background(), &tipoNovedad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(tipoNovedad)
}

// obtener todas las novedades
func GetNovedadesAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	return c.JSON(novedades)
}

// borrar novedad por id
func DeleteNovedad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("novedad eliminada")
}

// Actividades
// insertar actividad
func InsertActividad(c *fiber.Ctx) error {
	actividad := new(Actividades)
	if err := c.BodyParser(actividad); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("actividades")
	result, err := coll.InsertOne(context.TODO(), actividad)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(actividad)
}

// obtener actividad por id
func GetActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(actividad)
}

// obtener todas las actividades
func GetActividadAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(actividad)
}

// borrar actividad por id
func DeleteActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("actividad eliminada")
}

// Proveedores
// insertar proveedor
func InsertProveedor(c *fiber.Ctx) error {
	proveedor := new(Proveedores)
	if err := c.BodyParser(proveedor); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	result, err := coll.InsertOne(context.TODO(), proveedor)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(proveedor)
}

// obtener proveedor por id
func GetProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		fmt.Print(err)
	}
	return c.JSON(proveedor)
}

// obtener todos los proveedores
func GetProveedorAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		fmt.Print(err)
	}
	return c.JSON(proveedor)
}

// borrar proveedor por id
func DeleteProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("proveedor eliminado")
}

// obtener todos los centros de costos
func GetCecosAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		fmt.Print(err)
	}
	return c.JSON(cecos)
}

// obtener centro de costos por id
func GetCecos(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		fmt.Print(err)
	}
	return c.JSON(cecos)
}

// usuarios
// insertar usuario
func CreateUser(c *fiber.Ctx) error {
	newUser := new(User)
	if err := c.BodyParser(newUser); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	var user User
	dbUser.Where("email = ?", newUser.Email).First(&user)
	if user.User != "" {
		return c.SendString("el usuario ya existe")
	}

	if newUser.User == "" {
		newUser.User = strings.Split(newUser.Email, "@")[0]
	}
	newUser.Password = getMD5Hash(newUser.Password)
	dbUser.Create(&newUser)

	return c.Status(201).JSON(newUser)

}

// obtener usuario por user
func GetUser(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if !validateToken(token) {
		return c.SendString("token invalido")
	}

	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)

	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	return c.JSON(user)
}

// borrar usuario por user
func DeleteUser(c *fiber.Ctx) error {
	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	dbUser.Where("user = ?", item).Delete(&user)
	return c.SendString("usuario eliminado")
}

// actualizar usuario por user
func UpdateUser(c *fiber.Ctx) error {
	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	updateUser := new(User)
	if err := c.BodyParser(updateUser); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	newUser := user
	if updateUser.Email != "" {
		newUser.Email = updateUser.Email
	}
	if updateUser.Nombre != "" {
		newUser.Nombre = updateUser.Nombre
	}
	if updateUser.Apellido != "" {
		newUser.Apellido = updateUser.Apellido
	}
	if updateUser.Password != "" {
		newUser.Password = updateUser.Password
	}
	dbUser.Model(&user).Updates(newUser)
	return c.JSON(user)
}

// login
// logeo
func Login(c *fiber.Ctx) error {
	login := new(User)
	if err := c.BodyParser(login); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	var user User
	dbUser.Where("email = ?", login.Email).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	if user.Password != getMD5Hash(login.Password) {
		return c.SendString("contraseña incorrecta")
	}
	token := generateToken(32)
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
	})
	var token_auth Token_auth
	dbUser.Where("user_id = ?", user.UserId).First(&token_auth)
	if token_auth.TokenId == 0 {
		token_auth.Token = token
		token_auth.UserId = user.UserId
		token_auth.Generacion = time.Now().Format("2006-01-02 15:04:05")
		token_auth.Expiracion = time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
		dbUser.Create(&token_auth)
	} else {
		token_auth.Token = token
		token_auth.Generacion = time.Now().Format("2006-01-02 15:04:05")
		token_auth.Expiracion = time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
		dbUser.Model(&token_auth).Updates(token_auth)
	}
	return c.JSON(fiber.Map{
		"token": token,
		"user":  user.User,
	})
}

// token
// validar token
func validateToken(token string) bool {
	var token_auth Token_auth
	dbUser.Where("token = ?", token).First(&token_auth)
	if token_auth.TokenId == 0 {
		return false
	}
	if token_auth.Expiracion < time.Now().Format("2006-01-02 15:04:05") {
		return false
	}
	return true
}

// generar token
func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// md5
func getMD5Hash(message string) string {
	hash := md5.Sum([]byte(message))
	return hex.EncodeToString(hash[:])
}
