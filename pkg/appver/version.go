package appver

type Info struct {
	FilePath        string `json:"file_path"`
	CompanyName     string `json:"company_name"`
	FileDescription string `json:"file_description"`
	Version         int    `json:"version"`
	FullVersion     string `json:"full_version"`
	LegalCopyright  string `json:"legal_copyright"`
	ProductName     string `json:"product_name"`
	ProductVersion  string `json:"product_version"`
}

func New(filePath string) (*Info, error) {
	i := &Info{
		FilePath: filePath,
	}

	err := i.initialize()
	if err != nil {
		return nil, err
	}

	return i, nil
}
