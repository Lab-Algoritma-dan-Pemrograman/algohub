package entity

// Kolom NULL menjadi string kosong karena kode lama hanya memakai nilai ini lewat `a || b`.
type SSOUser struct {
	NIM       string
	Username  string
	FullName  string
	Nama      string
	Role      string
	ClassCode string
	Major     string
}

type ScheduleClass struct {
	ClassCode string
	Major     string
}
