package box

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	UserStatusActive                 = "active"
	UserStatusInactive               = "inactive"
	UserStatusCannotDeleteEdit       = "cannot_delete_edit"
	UserStatusCannotDeleteEditUpload = "cannot_delete_edit_upload"
)

type UsersResponse struct {
	TotalCount int         `json:"total_count"`
	Entries    []UserEntry `json:"entries"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

type UserEntry struct {
	Type          string `json:"type,omitempty"`
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	Login         string `json:"login,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	ModifiedAt    string `json:"modified_at,omitempty"`
	Language      string `json:"language,omitempty"`
	Timezone      string `json:"timezone,omitempty"`
	SpaceAmount   int64  `json:"space_amount,omitempty"`
	SpaceUsed     int    `json:"space_used,omitempty"`
	MaxUploadSize int64  `json:"max_upload_size,omitempty"`
	Status        string `json:"status,omitempty"`
	JobTitle      string `json:"job_title,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Address       string `json:"address,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
}

func (c *Client) UsersSearchAll(filterTerm string) ([]UserEntry, error) {
	// TODO: add method paramter for field list
	// TODO: add method paramter for user_type

	ues := []UserEntry{}

	offset := 0
	limit := 500

	// Get all users, looping through API pages
	for true {
		Url, err := url.Parse(fmt.Sprintf("%s/%s", c.APIBaseURL, "users"))
		if err != nil {
			return ues, err
		}
		parameters := url.Values{}
		parameters.Add("user_type", "all") // May be unnecessary
		// parameters.Add("fields", "id,name,login,status")
		parameters.Add("offset", fmt.Sprintf("%d", offset))
		parameters.Add("limit", fmt.Sprintf("%d", limit))
		parameters.Add("filter_term", filterTerm)
		Url.RawQuery = parameters.Encode()
		fmt.Println(Url.String())

		req, err := http.NewRequest("GET", Url.String(), nil)
		if err != nil {
			return ues, err
		}

		// make request with valid access token
		resp, err := c.HttpDo(req)
		if err != nil {
			return ues, err
		}

		if resp.StatusCode != http.StatusOK {
			return ues, fmt.Errorf("Unexpected status code while executing API request: %v", resp.Status)
		}

		// Read the response body
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		resp.Body.Close()
		// fmt.Println(buf.String())

		var ur UsersResponse
		if err := json.Unmarshal(buf.Bytes(), &ur); err != nil {
			return ues, err
		}
		// spew.Dump(ur)
		// fmt.Println(ur.TotalCount)

		ues = append(ues, ur.Entries...)

		// Use the values returned by the API response, not values passed in request
		offset = ur.Offset + ur.Limit

		if offset >= ur.TotalCount {
			break
		}
	}

	return ues, nil
}

func (c *Client) UsersGetAll() ([]UserEntry, error) {
	// TODO: add method paramter for field list
	// TODO: add method paramter for user_type

	ues := []UserEntry{}

	offset := 0
	limit := 500

	// Get all users, looping through API pages
	for true {
		Url, err := url.Parse(fmt.Sprintf("%s/%s", c.APIBaseURL, "users"))
		if err != nil {
			return ues, err
		}
		parameters := url.Values{}
		parameters.Add("user_type", "all") // May be unnecessary
		parameters.Add("fields", "id,name,login,status")
		parameters.Add("offset", fmt.Sprintf("%d", offset))
		parameters.Add("limit", fmt.Sprintf("%d", limit))
		Url.RawQuery = parameters.Encode()
		fmt.Println(Url.String())

		req, err := http.NewRequest("GET", Url.String(), nil)
		if err != nil {
			return ues, err
		}

		// make request with valid access token
		resp, err := c.HttpDo(req)
		if err != nil {
			return ues, err
		}

		if resp.StatusCode != http.StatusOK {
			return ues, fmt.Errorf("Unexpected status code while executing API request: %v", resp.Status)
		}

		// Read the response body
		buf := new(bytes.Buffer)
		io.Copy(buf, resp.Body)
		resp.Body.Close()
		// fmt.Println(buf.String())

		var ur UsersResponse
		if err := json.Unmarshal(buf.Bytes(), &ur); err != nil {
			return ues, err
		}
		// spew.Dump(ur)
		// fmt.Println(ur.TotalCount)

		ues = append(ues, ur.Entries...)

		// Use the values returned by the API response, not values passed in request
		offset = ur.Offset + ur.Limit

		if offset >= ur.TotalCount {
			break
		}
	}

	return ues, nil
}

func (c *Client) UsersGetUser(userID string) (UserEntry, error) {
	// TODO: add method paramter for field list
	// TODO: add method paramter for user_type

	ue := UserEntry{}

	Url, err := url.Parse(fmt.Sprintf("%s/%s/%s", c.APIBaseURL, "users", userID))
	if err != nil {
		return ue, err
	}
	parameters := url.Values{}
	// parameters.Add("fields", "id,name,login,status")
	Url.RawQuery = parameters.Encode()
	fmt.Println(Url.String())

	req, err := http.NewRequest("GET", Url.String(), nil)
	if err != nil {
		return ue, err
	}

	// make request with valid access token
	resp, err := c.HttpDo(req)
	if err != nil {
		return ue, err
	}

	if resp.StatusCode != http.StatusOK {
		return ue, fmt.Errorf("Unexpected status code while executing API request: %v", resp.Status)
	}

	// Read the response body
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	resp.Body.Close()
	// fmt.Println(buf.String())

	if err := json.Unmarshal(buf.Bytes(), &ue); err != nil {
		return ue, err
	}

	return ue, nil
}

func (c *Client) UsersUpdateUser(userID string, u *UserEntry) (*UserEntry, error) {
	// TODO: add method paramter for field list

	if userID == "" {
		return nil, errors.New("No userID provided")
	}
	if u == nil {
		return nil, errors.New("No updated UserEntry provided")
	}

	// Whitelist: remove un-settable attributes
	u.Type = ""
	u.ID = ""
	u.Login = ""
	u.CreatedAt = ""
	u.ModifiedAt = ""
	u.SpaceUsed = 0
	u.MaxUploadSize = 0
	u.AvatarURL = ""
	if !stringInSlice(u.Status, []string{UserStatusActive, UserStatusInactive, UserStatusCannotDeleteEdit, UserStatusCannotDeleteEditUpload}) {
		u.Status = ""
	}

	js, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(js))

	Url, err := url.Parse(fmt.Sprintf("%s/%s/%s", c.APIBaseURL, "users", userID))
	if err != nil {
		return nil, err
	}
	parameters := url.Values{}
	// parameters.Add("fields", "id,name,login,status")
	Url.RawQuery = parameters.Encode()
	fmt.Println(Url.String())

	req, err := http.NewRequest("PUT", Url.String(), bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	// make request with valid access token
	resp, err := c.HttpDo(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code while executing API request: %v", resp.Status)
	}

	// Read the response body
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	resp.Body.Close()
	// fmt.Println(buf.String())

	var ue UserEntry
	if err := json.Unmarshal(buf.Bytes(), &ue); err != nil {
		return nil, err
	}

	return &ue, nil
}
