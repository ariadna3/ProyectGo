package constantes

const (
	AdminRequired    = true
	AdminNotRequired = false
	AnyRol           = ""
	Admin            = "admin"
	PeopleOperation  = "po"
	Board            = "board"
	BoardAndPO       = "board,po"
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
const CollectionFreelance = "freelance"
const CollectionHistorial = "historial"

const (
	SheetGeneral    = "General"
	SheetHorasExtra = "Horas extra"
	SheetLicencias  = "Licencias"
)

const (
	PestanaGeneral         = "General"
	PestanaLicencias       = "Licencias"
	PestanaHorasExtras     = "Horas extras"
	PestanaNovedades       = "Novedades"
	PestanaPagoProvedores  = "PProvedores"
	PestanaFactServicios   = "Fact. de Servicios"
	PestanaRendGastos      = "Rendicion de gastos"
	PestanaNuevoCeco       = "Nuevo cecos"
	PestanaFactHorasExtras = "Fact. de Horas extras"
)

const (
	Periodo = "periodo"
	Fecha   = "fecha"
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
	DescTarjetaBeneficios = "Tarjeta de beneficios"
	DescPagos             = "Pagos"
	DescFactServicios     = "Facturacion de servicio"
	DescRendGastos        = "Rendicion de gastos"
	DescNuevoCeco         = "Nuevo ceco"
)

var HorasExtrasTipos = map[string]string{
	"50%diurno":       "G",
	"100%diurno":      "H",
	"50%nocturno":     "J",
	"100%nocturno":    "K",
	"feriadonocturno": "I",
	"feriadodiurno":   "I",
	"feriado":         "I",
}

var Permisos = map[string][]int{
	"admin":           []int{1, 2, 3},
	"po":              []int{1, 2, 3},
	"servicios":       []int{1, 2, 3},
	"ta":              []int{1, 2, 3},
	"comercial":       []int{1, 2, 3},
	"comunicaciones":  []int{1, 2, 3},
	"marketing":       []int{1, 2, 3},
	"externos":        []int{1, 2, 3},
	"board":           []int{1, 2, 3},
	"sustentabilidad": []int{1, 2, 3},
	"cultura":         []int{1, 2, 3},
	"comercialSop":    []int{1, 2, 3},
	"serviciosSop":    []int{1, 2, 3},
	"":                []int{1, 2, 3},
}

const (
	AceptarCambiosRecursos = "gerente,mail,fechaString"
)

const (
	DescripcionLicenciasComunes     = "Licencia comun"
	DescripcionLicenciasPatagonians = "Licencia Patagonians"
	DescripcionLicenciasOtras       = "Licencia otras"
)
