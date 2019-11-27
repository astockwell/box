package box

import (
	"reflect"
	"testing"
	"time"
)

func TestClient_FileUploadFromPath(t *testing.T) {
	type fields struct {
		ClientID                 string
		clientSecret             string
		EnterpriseID             string
		JWTKeyID                 string
		RSAPrivateKeyPemFilePath string
		GrantType                string
		APIBaseURL               string
		lastToken                *OauthTokenResponse
		lastTokenRetrieved       *time.Time
	}
	type args struct {
		localFilepath string
		boxFolderID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *FileUploadResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ClientID:                 tt.fields.ClientID,
				clientSecret:             tt.fields.clientSecret,
				EnterpriseID:             tt.fields.EnterpriseID,
				JWTKeyID:                 tt.fields.JWTKeyID,
				RSAPrivateKeyPemFilePath: tt.fields.RSAPrivateKeyPemFilePath,
				GrantType:                tt.fields.GrantType,
				APIBaseURL:               tt.fields.APIBaseURL,
				lastToken:                tt.fields.lastToken,
				lastTokenRetrieved:       tt.fields.lastTokenRetrieved,
			}
			got, err := c.FileUploadFromPath(tt.args.localFilepath, tt.args.boxFolderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.FileUploadFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.FileUploadFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
