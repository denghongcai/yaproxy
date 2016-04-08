package util

func BuildURL(host string, port int) (url string) {
	url = "tcp://" + host
	if port == 80 {
		url = "http://" + host
	} else if port == 443 {
		url = "https://" + host
	}
	return
}
