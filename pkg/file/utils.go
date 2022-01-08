package file

import "net/url"

//解析backupURl格式
// s3://my-bucket/my-dir/my-object.db
// s3 my-bucket my-dir/my-object.db
func ParseBackupURL(backupURL string) (string, string, string, error) {
	u, err := url.Parse(backupURL)
	if err != nil {
		return "", "", "", err
	}
	// s3,
	return u.Scheme, u.Host, u.Path[1:], nil
}
