package outline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func CheckUserExists(userID, userName, OutlineServer, OutlineToken string) (string, error) {
	fmt.Println("CheckUserExists")
	checkUserURL := fmt.Sprintf("%s/api/user/%s_%s", OutlineServer, userName, userID)

	log.Println(checkUserURL)

	req, err := http.NewRequest("GET", checkUserURL, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания GET-запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+OutlineToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения GET-запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа сервера: %v", err)
	}

	fmt.Println(resp.StatusCode)

	if resp.StatusCode == http.StatusOK {

		var getResponse struct {
			Links []string `json:"links"`
		}
		err := json.Unmarshal(body, &getResponse)
		if err != nil {
			return "", fmt.Errorf("ошибка разбора JSON: %v", err)
		}

		fmt.Println("body" + string(body))

		if len(getResponse.Links) > 0 {
			return getResponse.Links[0], nil
		}
		return "", nil
	} else if resp.StatusCode == http.StatusNotFound {

		var notFoundResponse struct {
			Detail string `json:"detail"`
		}
		err := json.Unmarshal(body, &notFoundResponse)
		if err != nil || notFoundResponse.Detail != "User not found" {
			return "", fmt.Errorf("неожиданный ответ сервера: %s", string(body))
		}
		return "", nil
	}

	return "", fmt.Errorf("неожиданный статус сервера: %d", resp.StatusCode)
}

func CreateNewUser(userId, userName, OutlineServer, OutlineToken string) (string, error) {
	requestBody := map[string]interface{}{
		"username": userName + "_" + userId,
		"proxies": map[string]interface{}{
			"shadowsocks": map[string]string{
				"method": "chacha20-ietf-poly1305",
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/user", OutlineServer), bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ошибка создания POST-запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+OutlineToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения POST-запроса: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка сервера: %d\n%s", resp.StatusCode, string(body))
	}

	var getResponse struct {
		Links []string `json:"links"`
	}
	err = json.Unmarshal(body, &getResponse)
	if err != nil {
		return "", fmt.Errorf("ошибка разбора JSON: %v", err)
	}

	fmt.Println("body" + string(body))

	if len(getResponse.Links) > 0 {
		return getResponse.Links[0], nil
	}
	return "", nil
}
