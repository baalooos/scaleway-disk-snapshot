# scaleway-disk-snapshot

Go program to snapshot and export to S3 Scaleway disk.

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
| PROJECT_ID | yes  | Your project ID|
| DEFAULT_REGION | yes  | The default region to use. Values are fr-par, nl-ams, pl-waw |
| DEFAULT_AZ | yes  | The default Az to use |
| DISK_ID  | yes  | ID of the disk to snapshot|
| SNAPSHOT_NUMBER  | yes  | Number of snapshots to keep|
| EXPORT_TO_S3  | yes  | Do your want to export the snapshot to S3? Value is true/false|
| BUCKET_NAME  | no  | Name of the bucket wwhere you want to store your snapshot |

For more informations on Scaleway Zones and Region, please refer to: [https://registry.terraform.io/providers/scaleway/scaleway/2.38.1/docs/guides/regions_and_zones](Scaleway Regions)
