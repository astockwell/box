package box

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type FileUploadRequest struct {
	Name   string                  `json:"name,omitempty"`
	Parent FileUploadRequestParent `json:"parent,omitempty"`
}
type FileUploadRequestParent struct {
	ID string `json:"id,omitempty"`
}

type FileUploadResponse struct {
	TotalCount int `json:"total_count"`
	Status     int `json:"status"`
	Entries    []struct {
		Type        `json:"type"`
		ID          `json:"id"`
		FileVersion struct {
			Type `json:"type"`
			ID   `json:"id"`
			Sha1 `json:"sha1"`
		} `json:"file_version"`
		SequenceID     `json:"sequence_id"`
		Etag           `json:"etag"`
		Sha1           `json:"sha1"`
		Name           `json:"name"`
		Description    `json:"description"`
		Size           `json:"size"`
		PathCollection struct {
			TotalCount `json:"total_count"`
			Entries    []struct {
				Type       `json:"type"`
				ID         `json:"id"`
				SequenceID `json:"sequence_id"`
				Etag       `json:"etag"`
				Name       `json:"name"`
			} `json:"entries"`
		} `json:"path_collection"`
		CreatedAt         `json:"created_at"`
		ModifiedAt        `json:"modified_at"`
		TrashedAt         `json:"trashed_at"`
		PurgedAt          `json:"purged_at"`
		ContentCreatedAt  `json:"content_created_at"`
		ContentModifiedAt `json:"content_modified_at"`
		CreatedBy         struct {
			Type  `json:"type"`
			ID    `json:"id"`
			Name  `json:"name"`
			Login `json:"login"`
		} `json:"created_by"`
		ModifiedBy struct {
			Type  `json:"type"`
			ID    `json:"id"`
			Name  `json:"name"`
			Login `json:"login"`
		} `json:"modified_by"`
		OwnedBy struct {
			Type  `json:"type"`
			ID    `json:"id"`
			Name  `json:"name"`
			Login `json:"login"`
		} `json:"owned_by"`
		SharedLink `json:"shared_link"`
		Parent     struct {
			Type       `json:"type"`
			ID         `json:"id"`
			SequenceID `json:"sequence_id"`
			Etag       `json:"etag"`
			Name       `json:"name"`
		} `json:"parent"`
		ItemStatus `json:"item_status"`
	} `json:"entries"`
}

type FileUploadResponseError struct {
	Type        string `json:"type"`
	Status      int    `json:"status"`
	Code        string `json:"code"`
	ContextInfo struct {
		Conflicts struct {
			Type        `json:"type"`
			ID          `json:"id"`
			FileVersion struct {
				Type `json:"type"`
				ID   `json:"id"`
				Sha1 `json:"sha1"`
			} `json:"file_version"`
			SequenceID `json:"sequence_id"`
			Etag       `json:"etag"`
			Sha1       `json:"sha1"`
			Name       `json:"name"`
		} `json:"conflicts"`
	} `json:"context_info"`
	HelpURL   string `json:"help_url"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (c *Client) FileUploadFromPath(localFilepath, boxFolderID string) (*FileUploadResponse, *FileUploadResponseError) {
	// Validation
	if localFilepath == "" {
		return nil, errors.New("No localFilepath provided")
	}
	if boxFolderID == "" {
		return nil, errors.New("No boxFolderID provided")
	}

	// filename := filepath.Base(localFilepath)

	Url, err := url.Parse(fmt.Sprintf("%s/%s", c.UploadBaseURL, "files/content"))
	if err != nil {
		return nil, err
	}

	// Read upload file
	file, err := os.Open(localFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
	)

	// write the file
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	// write the other form fields we need
	fureq := FileUploadRequest{
		Name: fi.Name(),
		Parent: FileUploadRequestParent{
			ID: boxFolderID,
		},
	}
	js, err := json.Marshal(&fureq)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("attributes", string(js))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", Url.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	// make request with valid access token
	resp, err := c.HttpDo(req)
	if err != nil {
		return nil, err
	}

	// Read the response body
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	resp.Body.Close()
	// fmt.Println(buf.String())

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		var fure FileUploadResponseError
		if err := json.Unmarshal(buf.Bytes(), &fure); err != nil {
			return nil, fmt.Errorf("Error json.Unmarshal(&fure): %v. Body: %v", err, buf.String())
		}
		return nil, &fure
	}

	var fur FileUploadResponse
	if err := json.Unmarshal(buf.Bytes(), &fur); err != nil {
		return nil, fmt.Errorf("Error json.Unmarshal(&fur): %v. Body: %v", err, buf.String())
	}

	// Add status code for later inspection
	fur.Status = resp.StatusCode

	return &fur, nil
}

func (c *Client) FileUploadVersionFromPath(localFilepath, boxFileID string) (*FileUploadResponse, error) {
	// Validation
	if localFilepath == "" {
		return nil, errors.New("No localFilepath provided")
	}
	if boxFileID == "" {
		return nil, errors.New("No boxFileID provided")
	}

	Url, err := url.Parse(fmt.Sprintf("%s/files/%s/content", c.UploadBaseURL, boxFileID))
	if err != nil {
		return nil, err
	}

	// Read upload file
	file, err := os.Open(localFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
	)

	// write the file
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	// write the other form fields we need
	fureq := FileUploadRequest{
		Name: fi.Name(),
	}
	js, err := json.Marshal(&fureq)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("attributes", string(js))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", Url.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	// make request with valid access token
	resp, err := c.HttpDo(req)
	if err != nil {
		return nil, err
	}

	// Read the response body
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	resp.Body.Close()
	// fmt.Println(buf.String())

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code while executing API request: %v. Body: %v", resp.Status, buf.String())
	}

	var fur FileUploadResponse
	if err := json.Unmarshal(buf.Bytes(), &fur); err != nil {
		return nil, err
	}

	return &fur, nil
}

func (c *Client) FileDownload(boxFileID string) (*http.Response, error) {
	if boxFileID == "" {
		return nil, errors.New("No boxFileID provided")
	}

	Url, err := url.Parse(fmt.Sprintf("%s/files/%s/content", c.APIBaseURL, boxFileID))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", Url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpDo(req)
	if err != nil {
		log.Fatal(err)
	}

	// spew.Dump(resp.StatusCode)

	// // Read the response body
	// buf := new(bytes.Buffer)
	// io.Copy(buf, resp.Body)
	// resp.Body.Close()
	// fmt.Println(buf.String())

	return resp, nil
}

func (c *Client) FileDownloadGetContent(boxFileID string) (*bytes.Buffer, error) {
	resp, err := c.FileDownload(boxFileID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP non-200 status: %v (must manually handle via c.FileDownload() )", resp.StatusCode)
	}

	// Read the response body
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	resp.Body.Close()

	return buf, nil
}
