package app

type Config struct {
	API struct {
		Listen string `yaml:"listen"`
		DbDsn  string `yaml:"dbDsn"`
	} `yaml:"api"`
}
