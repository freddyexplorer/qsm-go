package m3util

import (
	"encoding/csv"
	"log"
	"os"
)

func ChangeToDocsGeneratedDir() {
	changeToDocsSubdir("generated")
}

func ChangeToDocsDataDir() {
	changeToDocsSubdir("data")
}

func changeToDocsSubdir(subDir string) {
	if _, err := os.Stat("docs"); !os.IsNotExist(err) {
		ExitOnError(os.Chdir("docs"))
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			ExitOnError(os.Mkdir(subDir, os.ModePerm))
		}
		ExitOnError(os.Chdir(subDir))
	}
}

func ExitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/*func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	ExitOnError(err)
}
*/

func CloseFile(file *os.File) {
	ExitOnError(file.Close())
}

func WriteNextString(file *os.File, text string) {
	_, err := file.WriteString(text)
	ExitOnError(err)
}

func WriteAll(writer *csv.Writer, records [][]string) {
	ExitOnError(writer.WriteAll(records))
}

func Write(writer *csv.Writer, records []string) {
	ExitOnError(writer.Write(records))
}
