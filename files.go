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
	Entries    []struct {
		ID             int    `json:"id"`
		Etag           int    `json:"etag"`
		Type           string `json:"type"`
		SequenceID     int    `json:"sequence_id"`
		Name           string `json:"name"`
		Sha1           string `json:"sha1"`
		Description    string `json:"description"`
		Size           int    `json:"size"`
		PathCollection struct {
			TotalCount int `json:"total_count"`
			Entries    []struct {
				ID         int    `json:"id"`
				Etag       int    `json:"etag"`
				Type       string `json:"type"`
				SequenceID int    `json:"sequence_id"`
				Name       string `json:"name"`
			} `json:"entries"`
		} `json:"path_collection"`
		CreatedAt         string `json:"created_at"`
		ModifiedAt        string `json:"modified_at"`
		TrashedAt         string `json:"trashed_at"`
		PurgedAt          string `json:"purged_at"`
		ContentCreatedAt  string `json:"content_created_at"`
		ContentModifiedAt string `json:"content_modified_at"`
		CreatedBy         struct {
			ID    int    `json:"id"`
			Type  string `json:"type"`
			Name  string `json:"name"`
			Login string `json:"login"`
		} `json:"created_by"`
		ModifiedBy struct {
			ID    int    `json:"id"`
			Type  string `json:"type"`
			Name  string `json:"name"`
			Login string `json:"login"`
		} `json:"modified_by"`
		OwnedBy struct {
			ID    int    `json:"id"`
			Type  string `json:"type"`
			Name  string `json:"name"`
			Login string `json:"login"`
		} `json:"owned_by"`
		SharedLink struct {
			URL                 string `json:"url"`
			DownloadURL         string `json:"download_url"`
			VanityURL           string `json:"vanity_url"`
			Access              string `json:"access"`
			EffectiveAccess     string `json:"effective_access"`
			EffectivePermission string `json:"effective_permission"`
			UnsharedAt          string `json:"unshared_at"`
			IsPasswordEnabled   bool   `json:"is_password_enabled"`
			Permissions         struct {
				CanDownload bool `json:"can_download"`
				CanPreview  bool `json:"can_preview"`
			} `json:"permissions"`
			DownloadCount int `json:"download_count"`
			PreviewCount  int `json:"preview_count"`
		} `json:"shared_link"`
		Parent struct {
			ID         int    `json:"id"`
			Etag       int    `json:"etag"`
			Type       string `json:"type"`
			SequenceID int    `json:"sequence_id"`
			Name       string `json:"name"`
		} `json:"parent"`
		ItemStatus    string `json:"item_status"`
		VersionNumber int    `json:"version_number"`
		CommentCount  int    `json:"comment_count"`
		Permissions   struct {
			CanDelete              bool `json:"can_delete"`
			CanDownload            bool `json:"can_download"`
			CanInviteCollaborator  bool `json:"can_invite_collaborator"`
			CanRename              bool `json:"can_rename"`
			CanSetShareAccess      bool `json:"can_set_share_access"`
			CanShare               bool `json:"can_share"`
			CanAnnotate            bool `json:"can_annotate"`
			CanComment             bool `json:"can_comment"`
			CanPreview             bool `json:"can_preview"`
			CanUpload              bool `json:"can_upload"`
			CanViewAnnotationsAll  bool `json:"can_view_annotations_all"`
			CanViewAnnotationsSelf bool `json:"can_view_annotations_self"`
		} `json:"permissions"`
		Tags []string `json:"tags"`
		Lock struct {
			ID        int    `json:"id"`
			Type      string `json:"type"`
			CreatedBy struct {
				ID    int    `json:"id"`
				Type  string `json:"type"`
				Name  string `json:"name"`
				Login string `json:"login"`
			} `json:"created_by"`
			CreatedAt           string `json:"created_at"`
			ExpiredAt           string `json:"expired_at"`
			IsDownloadPrevented bool   `json:"is_download_prevented"`
		} `json:"lock"`
		Extension         string `json:"extension"`
		IsPackage         bool   `json:"is_package"`
		ExpiringEmbedLink struct {
			AccessToken  string `json:"access_token"`
			ExpiresIn    int    `json:"expires_in"`
			TokenType    string `json:"token_type"`
			RestrictedTo []struct {
				Scope  string `json:"scope"`
				Object struct {
					ID         int    `json:"id"`
					Etag       int    `json:"etag"`
					Type       string `json:"type"`
					SequenceID int    `json:"sequence_id"`
					Name       string `json:"name"`
				} `json:"object"`
			} `json:"restricted_to"`
			URL string `json:"url"`
		} `json:"expiring_embed_link"`
		WatermarkInfo struct {
			IsWatermarked bool `json:"is_watermarked"`
		} `json:"watermark_info"`
		AllowedInviteeRoles []string `json:"allowed_invitee_roles"`
		IsExternallyOwned   bool     `json:"is_externally_owned"`
		HasCollaborations   bool     `json:"has_collaborations"`
		Metadata            struct {
			Global struct {
				MarketingCollateral struct {
					CanEdit     bool   `json:"$canEdit"`
					ID          string `json:"$id"`
					Parent      string `json:"$parent"`
					Scope       string `json:"$scope"`
					Template    string `json:"$template"`
					Type        string `json:"$type"`
					TypeVersion int    `json:"$typeVersion"`
					Version     int    `json:"$version"`
				} `json:"marketingCollateral"`
			} `json:"global"`
		} `json:"metadata"`
		ExpiresAt       string `json:"expires_at"`
		Representations struct {
			Entries []struct {
				Content struct {
					URLTemplate string `json:"url_template"`
				} `json:"content"`
				Info struct {
					URL string `json:"url"`
				} `json:"info"`
				Properties struct {
					Dimensions string `json:"dimensions"`
					Paged      bool   `json:"paged"`
					Thumb      bool   `json:"thumb"`
				} `json:"properties"`
				Representation string `json:"representation"`
				Status         struct {
					State string `json:"state"`
				} `json:"status"`
			} `json:"entries"`
		} `json:"representations"`
	} `json:"entries"`
}

func (c *Client) FileUploadFromPath(localFilepath, boxFolderID string) (*FileUploadResponse, error) {
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

	if resp.StatusCode != http.StatusOK || resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Unexpected status code while executing API request: %v. Body: %v", resp.Status, buf.String())
	}

	var fur FileUploadResponse
	if err := json.Unmarshal(buf.Bytes(), &fur); err != nil {
		return nil, err
	}

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
