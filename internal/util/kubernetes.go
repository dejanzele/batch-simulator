package util

func CreateImageString(image, tag string) string {
	if tag == "" {
		tag = "latest"
	}
	return image + ":" + tag
}
