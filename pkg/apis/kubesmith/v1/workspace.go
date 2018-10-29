package v1

type WorkspaceStorage struct {
	S3 WorkspaceStorageS3 `json:"s3"`
}

type WorkspaceStorageS3 struct {
	Host        string                        `json:"host"`
	Port        int                           `json:"path"`
	UseSSL      bool                          `json:"useSSL"`
	BucketName  string                        `json:"bucketName"`
	Credentials WorkspaceStorageS3Credentials `json:"credentials"`
}

type WorkspaceStorageS3Credentials struct {
	Secret WorkspaceStorageS3CredentialsS3 `json:"secret"`
}

type WorkspaceStorageS3CredentialsS3 struct {
	Name         string `json:"name"`
	AccessKeyKey string `json:"accessKeyKey"`
	SecretKeyKey string `json:"secretKeyKey"`
}

type WorkspaceRepo struct {
	URL string           `json:"url"`
	SSH WorkspaceRepoSSH `json:"ssh"`
}

type WorkspaceRepoSSH struct {
	Secret WorkspaceRepoSSHSecret `json:"secret"`
}

type WorkspaceRepoSSHSecret struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}
