package models

func BuildFilePath(collection string, id string, filename string) string {
	return "/" + CDN_BASE_ROUTE + "/" + FILE_ROUTE + "/" + collection + "/" + id + "/" + filename

}
