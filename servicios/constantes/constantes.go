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
	"50%diurno":       "E",
	"100%diurno":      "F",
	"50%nocturno":     "H",
	"100%nocturno":    "I",
	"feriadonocturno": "G",
	"feriadodiurno":   "G",
	"feriado":         "G",
}

const (
	AceptarCambiosRecursos = "gerente,mail,fechaString"
)
