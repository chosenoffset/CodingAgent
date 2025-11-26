package vector

type CodeSnippet struct {
	ID           string `json:"id"`
	FilePath     string `json:"file_path"`
	Language     string `json:"language"`
	FunctionName string `json:"function_name"`
	Code         string `json:"code"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	DocString    string `json:"doc_string"`
}
