package constantes

const (
	AdminRequired    = true
	AdminNotRequired = false
	AnyRol           = ""
	Admin            = "admin"
	PeopleOperation  = "po"
)

var AllRol string = ""

const FormatoFecha = "2006-02-01"

const Database = "portalDeNovedades"

const (
	CollectionPasosWorkflow = "pasosWorkflow"
	CollectionProveedor     = "proveedores"
	CollectionActividad     = "actividades"
	CollectionNovedad       = "novedades"
	CollectionRecurso       = "recursos"
	CollectionUserITP       = "usersITP"
	CollectionCecos         = "centroDeCostos"
)

const (
	PestanaGeneral     = "General"
	PestanaLicencias   = "Licencias"
	PestanaHorasExtras = "Horas extras"
)

const (
	Pendiente = "pendiente"
	Aceptada  = "aceptada"
	Rechazada = "rechazada"
)

const (
	TipoGerente = "manager"
	TipoGrupo   = "grupo"
)

const (
	DescSueldoNuevoMasivo = "Nuevo sueldo masivo"
	DescSueldoNuevo       = "Nuevo sueldo"
	DescHorasExtras       = "Horas extras"
	DescAnticipo          = "Anticipo"
	DescPrestamo          = "Prestamo"
	DescLicencia          = "Licencia"
	DescPagos             = "Pagos"
)

var HorasExtrasTipos = map[string]string{
	"50%diurno":       "F",
	"100%diurno":      "G",
	"50%nocturno":     "I",
	"100%nocturno":    "J",
	"feriadonocturno": "H",
	"feriadodiurno":   "H",
	"feriado":         "H",
}

const (
	AceptarCambiosRecursos = "gerente,mail,fechaString"
)
