package main

import (
	"fmt"

	"os"
	"strconv"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func priv_create_snapshot(disk_id string, snapshot_name string, project_id string, instanceApi *instance.API, zone scw.Zone) (string, error) {
	// Create Snapshot
	fmt.Println("Creating snapshot")
	create_snapshot, err := instanceApi.CreateSnapshot(&instance.CreateSnapshotRequest{
		Zone:     zone,
		Name:     snapshot_name,
		VolumeID: &disk_id,
		Project:  &project_id,
	})
	if err == nil {
		// Print Snapshot Name and ID
		fmt.Println("Snapshot ID")
		fmt.Println(create_snapshot.Snapshot.ID)
		fmt.Println("Snapshot name")
		fmt.Println(create_snapshot.Snapshot.Name)

		return create_snapshot.Snapshot.ID, err

	} else {
		return "", err
	}

}

func priv_wait_for_snapshot(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	// Wait for Snapshot
	fmt.Println("Waiting for snapshot")
	_, err := instanceApi.WaitForSnapshot(&instance.WaitForSnapshotRequest{
		SnapshotID: snapshot_id,
		Zone:       zone,
	})
	return err
}

func priv_export_to_s3(snapshot_id string, snapshot_name string, bucket string, instanceApi *instance.API, zone scw.Zone) error {
	// Export Snapshot to S3 Bucket
	fmt.Println("Exporting Snapshot")
	export_snapshot, err := instanceApi.ExportSnapshot(&instance.ExportSnapshotRequest{
		Zone:       scw.ZoneFrPar1,
		SnapshotID: snapshot_id,
		Bucket:     bucket,
		Key:        snapshot_name,
	})
	if err != nil {
		return err
	} else {
		fmt.Println("Export task ID")
		fmt.Println(export_snapshot.Task.ID)

		// Wait for export
		err = priv_wait_for_export(snapshot_id, instanceApi, zone)

		return err
	}
}

func priv_wait_for_export(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	fmt.Println("Waiting for Export")
	get_snapshot, err := instanceApi.GetSnapshot(&instance.GetSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshot_id,
	})
	fmt.Println(err)
	if err != nil {
		return err
	} else {

		for get_snapshot.Snapshot.State != "available" {
			fmt.Println("Snapshot Status")
			fmt.Println(get_snapshot.Snapshot.State)
			get_snapshot, err = instanceApi.GetSnapshot(&instance.GetSnapshotRequest{
				Zone:       zone,
				SnapshotID: snapshot_id,
			})
			if err != nil {
				return err
			}
			time.Sleep(30 * time.Second)
			fmt.Println("Waiting for Snapshot export")
		}

		return err
	}
}

func list_snapshot(disk_id string, instanceApi *instance.API, zone scw.Zone) (*instance.ListSnapshotsResponse, error) {
	fmt.Println("Get Snapshot List for Disk")

	list_snapshot, err := instanceApi.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone:         zone,
		BaseVolumeID: &disk_id,
	})

	return list_snapshot, err
}

func priv_delete_snapshot(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	fmt.Println("Deleting snapshot")
	err := instanceApi.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshot_id,
	})
	return err
}

func main() {

	// Get Env variables
	scw_access_key, validate := os.LookupEnv("SCW_ACCESS_KEY")
	if !validate {
		panic("You must set SCW_ACCESS_KEY")
	}
	scw_secret_key, validate := os.LookupEnv("SCW_SECRET_KEY")
	if !validate {
		panic("You must set SCW_SECRET_KEY")
	}
	organizationID, validate := os.LookupEnv("ORGANIZATION_ID")
	if !validate {
		panic("You must set ORGANIZATION_ID")
	}
	project_id, validate := os.LookupEnv("PROJECT_ID")
	if !validate {
		panic("You must set PROJECT_ID")
	}
	disk_id, validate := os.LookupEnv("DISK_ID")
	if !validate {
		panic("You must set DISK_ID")
	}

	string_snapshot_number, validate := os.LookupEnv("SNAPSHOT_NUMBER")
	if !validate && string_snapshot_number == "true" {
		panic("You must set SNAPSHOT_NUMBER")
	}
	snapshot_number, err := strconv.Atoi(string_snapshot_number)
	if err != nil {
		panic(err)
	}

	// Forge snapshot name
	now := time.Now()
	snapshot_name := fmt.Sprint("snapshot.", now.Unix())

	// Create a Scaleway client
	client, err := scw.NewClient(
		// Get your organization ID at https://console.scaleway.com/organization/settings
		scw.WithDefaultOrganizationID(organizationID),
		// Get your credentials at https://console.scaleway.com/iam/api-keys
		scw.WithAuth(scw_access_key, scw_secret_key),
		// Get more about our availability zones at https://www.scaleway.com/en/docs/console/my-account/reference-content/products-availability/
		scw.WithDefaultRegion("fr-par"),
	)
	if err != nil {
		panic(err)
	}

	// Create SDK objects for Scaleway Instance product
	instanceApi := instance.NewAPI(client)

	// Create Snapshot
	snapshot_id, err := priv_create_snapshot(disk_id, snapshot_name, project_id, instanceApi, scw.ZoneFrPar1)
	if err != nil {
		panic(err)
	}

	// Wait for Snapshot createion
	err = priv_wait_for_snapshot(snapshot_id, instanceApi, scw.ZoneFrPar1)
	if err != nil {
		panic(err)
	}

	export_to_s3, validate := os.LookupEnv("EXPORT_TO_S3")
	if !validate {
		panic("You must set EXPORT_TO_S3")
	}

	// Export Snapshot to S3
	if export_to_s3 == "true" {

		bucket, validate := os.LookupEnv("BUCKET_NAME")
		if !validate && export_to_s3 == "true" {
			panic("You must set BUCKET_NAME")
		}

		err = priv_export_to_s3(snapshot_id, snapshot_name, bucket, instanceApi, scw.ZoneFrPar1)
		if err != nil {
			panic(err)
		}
	}

	// LIST Snapshot
	snapshot_list, err := list_snapshot(disk_id, instanceApi, scw.ZoneFrPar1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Number of Snapshot")
	fmt.Println(snapshot_list.TotalCount)

	oldest := snapshot_list.Snapshots[0]
	for _, snap := range snapshot_list.Snapshots {
		if snap.CreationDate.Unix() < oldest.CreationDate.Unix() {
			oldest = snap
		}
	}

	if snapshot_list.TotalCount > uint32(snapshot_number) {
		// DELETE Snapshot
		fmt.Println(oldest.ID)
		err = priv_delete_snapshot(oldest.ID, instanceApi, scw.ZoneFrPar1)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("No Snapshot to delete")
	}
}
