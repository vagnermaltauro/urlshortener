package repository

type IDGenerator interface {
	Generate() (int64, error)

	Encode(id int64) string

	Decode(code string) (int64, error)
}
