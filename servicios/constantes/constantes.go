package constantes

const (
	AdminRequired    = true
	AdminNotRequired = false
	AnyRol           = ""
	Admin            = "admin"
	AllRol           = "admin,po,servicios,ta,comercial,comunicaciones,marketing,externos"
)

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
	DescSueldoNuevo = "Nuevo sueldo"
	DescHorasExtras = "Horas extras"
	DescAnticipo    = "Anticipo"
	DescPrestamo    = "Prestamo"
	DescLicencia    = "Licencia"
)

const (
	AceptarCambiosRecursos = "gerente,mail,fechaString"
)
