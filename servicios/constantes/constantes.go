package constantes

const (
	AdminRequired    = true
	AdminNotRequired = false
	AnyRol           = ""
	Admin            = "admin"
	PeopleOperation  = "po"
)

const CecosNotValids = "(99999999999)"

var AllRol string = ""

const FormatoFecha = "2006-02-01"
const FormatoFechaProvicional = "2006-01-02"

const Database = "portalDeNovedades"
const CollectionPasosWorkflow = "pasosWorkflow"
const CollectionProveedor = "proveedores"
const CollectionActividad = "actividades"
const CollectionNovedad = "novedades"
const CollectionRecurso = "recursos"
const CollectionUserITP = "usersITP"
const CollectionCecos = "centroDeCostos"

const (
	SheetGeneral    = "General"
	SheetHorasExtra = "Horas extra"
	SheetLicencias  = "Licencias"
)

const (
	PestanaGeneral     = "General"
	PestanaLicencias   = "Licencias"
	PestanaHorasExtras = "Horas extras"
	PestanaNovedades   = "Novedades"
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
