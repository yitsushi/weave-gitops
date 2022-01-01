package types

type Repo struct {
	Name            string
	GitopsSourceUrl string
	Applications    map[string]App
}
