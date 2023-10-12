package models

import (
	"time"
)

// ----Freelances----
type Freelances struct {
	IdFreelance     int       `bson:"idFreelance"`
	NroFreelance    int       `bson:"nroFreelance"`
	CUIT            string    `bson:"cuit"`
	Apellido        string    `bson:"apellido"`
	Nombre          string    `bson:"nombre"`
	FechaIngreso    time.Time `bson:"fechaIngreso"`
	Nomina          string    `bson:"nomina"`
	Gerente         int       `bson:"gerente"`
	Vertical        string    `bson:"vertical"`
	HorasMen        int       `bson:"horasMen"`
	Cargo           string    `bson:"cargo"`
	FacturaMonto    float64   `bson:"facturaMondo"`
	FacturaDesde    time.Time `bson:"facturaDesde"`
	FacturaADCuit   string    `bson:"facturaADCuit"`
	FacturaADMonto  float64   `bson:"facturaADMonto"`
	FacturaADDesde  time.Time `bson:"facturaADDesde"`
	B21Monto        float64   `bson:"b21Monto"`
	B21Desde        time.Time `bson:"b21Desde"`
	Comentario      string    `bson:"comentario"`
	Habilitado      string    `bson:"habilitado"`
	FechaBaja       time.Time `bson:"fechaBaja"`
	Cecos           []Rcc     `bson:"cecos"`
	Telefono        string    `bson:"telefono"`
	EmailLaboral    string    `bson:"emailLaboral"`
	EmailPersonal   string    `bson:"emailPersonal"`
	FechaNacimiento time.Time `bson:"fechaNacimiento"`
	Genero          string    `bson:"genero"`
	Nacionalidad    string    `bson:"nacionalidad"`
	DomCalle        string    `bson:"domCalle"`
	DomNumero       int       `bson:"domNumero"`
	DomPiso         int       `bson:"domPiso"`
	DomDepto        string    `bson:"domDepto"`
	DomLocalidad    string    `bson:"domLocalidad"`
	DomProvincia    string    `bson:"domProvincia"`
}

type Rcc struct {
	CcId         int     `bson:"ccId"`
	CcNum        int     `bson:"ccNum"`
	CcPorcentaje float32 `bson:"ccPorcentaje"`
	CcNombre     string  `bson:"ccNombre"`
	CcCliente    string  `bson:"ccCliente"`
}

// Historial de cambios freelance
type HistorialFreelance struct {
	IdHistorial     int                    `bson:"idHistorial"`
	UsuarioEmail    string                 `bson:"usuarioEmail"`
	UsuarioNombre   string                 `bson:"usuarioNombre"`
	UsuarioApellido string                 `bson:"usuarioApellido"`
	Freelance       map[string]interface{} `bson:"freelance"`
	Tipo            string                 `bson:"tipo"`
	FechaHora       time.Time              `bson:"fechaHora"`
}

// Historial de cambios staff
type HistorialStaff struct {
	IdHistorial     int                    `bson:"idHistorial"`
	UsuarioEmail    string                 `bson:"usuarioEmail"`
	UsuarioNombre   string                 `bson:"usuarioNombre"`
	UsuarioApellido string                 `bson:"usuarioApellido"`
	Staff           map[string]interface{} `bson:"staff"`
	Tipo            string                 `bson:"tipo"`
	FechaHora       time.Time              `bson:"fechaHora"`
}

type Staff struct {
	IdStaff            int       `bson:"idStaff"`
	NroLegajo          int       `bson:"nroLegajo"`
	CUIT               string    `bson:"cuit"`
	Apellido           string    `bson:"apellido"`
	Nombre             string    `bson:"nombre"`
	FechaIngreso       time.Time `bson:"fechaIngreso"`
	Nomina             string    `bson:"nomina"`
	Gerente            int       `bson:"gerente"`
	Vertical           string    `bson:"vertical"`
	HorasMen           int       `bson:"horasMen"`
	Cargo              string    `bson:"cargo"`
	SueldoFacturaMonto float64   `bson:"sueldoFacturaMonto"`
	FacturaDesde       time.Time `bson:"facturaDesde"`
	FacturaADCuit      string    `bson:"facturaADCuit"`
	FacturaADMonto     float64   `bson:"facturaADMonto"`
	FacturaADDesde     time.Time `bson:"facturaADDesde"`
	B21Monto           float64   `bson:"b21Monto"`
	B21Desde           time.Time `bson:"b21Desde"`
	Comentario         string    `bson:"comentario"`
	Habilitado         string    `bson:"habilitado"`
	FechaBaja          time.Time `bson:"fechaBaja"`
}
