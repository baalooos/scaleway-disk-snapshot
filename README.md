# scaleway-disk-snapshot

Go program to snashot and export to S3 Scaleway disk.

## Usage

- Clone this repository
- `go run .`

## Environment Variables

This program read your env variables and expect to find the following variables:

| Variable Name  | Mandatory | Description |
| ------------- | ------------- | ------------- |
| SCW_ACCESS_KEY  | yes  | Your Scaleway access key|
| SCW_SECRET_KEY  | yes  | Your Scaleway secret key|
| ORGANIZATION_ID  | yes  | Your organization ID|
| PROJECT_ID  | yes  | Your project ID|
| DISK_ID  | yes  | ID of the disk to snapshot|
| SNAPSHOT_NUMBER  | yes  | Number of snapshots to keep|
| EXPORT_TO_S3  | yes  | Do your want to export the snapshot to S3? Value is true/false|
| BUCKET_NAME  | no  | Name of the bucket wwhere you want to store your snapshot |
