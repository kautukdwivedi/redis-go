package internal

type RDBFile struct {
	Dir        string
	DBfilename string
}

func NewRDBFile(dir, dbfilename string) *RDBFile {
	return &RDBFile{
		Dir:        dir,
		DBfilename: dbfilename,
	}
}
