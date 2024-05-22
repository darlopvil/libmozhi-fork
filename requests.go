package libmozhi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func postRequest(url string, data []byte, contenttype string) (string, error) {
	bodyReader := bytes.NewReader(data)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return "", err
	}

	UserAgent := "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
	r.Header.Set("Content-Type", contenttype)
	r.Header.Set("User-Agent", UserAgent)
	r.Header.Set("Accept", "application/json, text/plain, */*")
	if strings.Contains(url, "translate.google.com") {
		r.Header["X-Goog-BatchExecute-Bgr"] = []string{`[";zdO404vQAAY15hn3FeZfsjrAbuuehnMmADQBEArZ1G-T9qS8JS9naDGjHzet_oFy3jZh71nxVmXUAikXE47SFgWlmuJyjXXQXss_cYlCHwAAADRPAAAACXUBB2MANh6IYRPOcvfOYUJskh9A8304dJA_u1Q7_ya6k8Z5quWDFm3VQiFjwsOFc7A1BrN57sTlGgX4gBcAGjiZ_U7QyQiF8546TcSILSCFIA9cRZA7hdQ-hAKS3L-4-Btym9_roHKYx4a2hKMfrWzakho2gD4J19lvIso4ot_KxS5R-hZ_IoA1vt4scD2mxm-zYypgeimoQ4Yep9SiLG-nqebyGG9FBAWZgMn8c06mzMHNqtxl_Q1z3E8KXknpQ6GKD8_mUpxpkYKsOsJCx9pxOcWU8ESjVgJhcnS8sAP33hX4GIKUC3R3Nca3HhoLTGLYaajWoI9Kaw1FlYkNKNPzWBo03k77hv8hsiAAJRMBhajAUaBRAzPL95qrpQITPKnXm9mqNx8fUcLq-6bYPOjwuUFEyClKf49UtK6d1ofIDVvqndOdSVxY_lDHvJmsEEpS4LL6xPCys1x1TZ6kxP5Rl--5YhG6s5BtQ1djOmQ4CG6qsRtNLT-fn2HOUH2yfmMzf4UGGQNVTHmnmpybIoZ7XWrRUrnjbc8HsRr6qE2i4n439o6hl7oc4LoslFKJgjn4OeS26LhYYp1VnOtZcWSd6PpsG8ypGqzgp4B82G_zkEL_Ag6U8U6UXR83ZHfBdbp0DFKDhnAdpS6snYp4g3CNMH4-ifLFvmVP22u-IIBXKogyY8Ov8txbZxngRpl7qqXG8ShSdyijM14IqLPVDjWVwkCyXMVLgiWH83EwzQ5_MmcrU1DdxZ_HKTotfY9RJpDijilj7MLbYDUTI0J8BSFqxqaDCZdXECCJ46UYGF1I7H29l9K8wy8aHvMRldqupGT_AN5n6GDkg2Zmca7elM624dzEqqW62uYiAaxznp7AAWQYO5NPe--VcnNm36x3KQ5Q2paQAx4-IqpVEHozNLhc2ma20yvFpQHJ2put8cvkxn8WHbBgs8cXhq3NnhfKbyZuJNIfAOqevqMcNJ8BTLx5j2tWw3HSQtbNpONwHw",null,null,31,null,null,null,0,"2"]`}
	}
	client := &http.Client{}
	res, err0 := client.Do(r)
	if err0 != nil {
		return "", err0
	}

	defer res.Body.Close()

	body, err1 := io.ReadAll(res.Body)
	if err1 != nil {
		return "", err1
	}

	return string(body), nil
}

func getRequest(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	UserAgent := "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
	// r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", UserAgent)

	client := &http.Client{}
	res, err0 := client.Do(r)
	if err0 != nil {
		return "", err0
	}

	defer res.Body.Close()

	body, err1 := io.ReadAll(res.Body)
	if err1 != nil {
		return "", err1
	}
	return string(body), nil
}
