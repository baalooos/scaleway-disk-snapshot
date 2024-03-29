package main

import (
	"errors"
	"fmt"
	"slices"

	"os"
	"strconv"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type snapshot_config struct {
	scw_access_key  string
	scw_secret_key  string
	organizationID  string
	project_id      string
	snapshot_region scw.Region
	snapshot_az     scw.Zone
	disk_id         string
	snapshot_number int
	export_to_s3    string
}

// Todo add a formated date function

func get_env_var() snapshot_config {
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
	default_region, validate := os.LookupEnv("DEFAULT_REGION")
	if !validate {
		panic("You must set DEFAULT_REGION")
	}
	//Get the region from the Env Variable
	snapshot_region, err := scw.ParseRegion(default_region)
	if err != nil {
		panic(err)
	}
	default_az, validate := os.LookupEnv("DEFAULT_AZ")
	if !validate {
		panic("You must set DEFAULT_AZ")
	}
	// Get the zone from the Env Variable
	snapshot_az, err := scw.ParseZone(default_az)
	if err != nil {
		panic(err)
	}
	// Check if our Az is in our Region
	if !slices.Contains(scw.Region.GetZones(snapshot_region), snapshot_az) {
		panic("The default_AZ must be in the Default_Region")
	}

	disk_id, validate := os.LookupEnv("DISK_ID")
	if !validate {
		panic("You must set DISK_ID")
	}
	string_snapshot_number, validate := os.LookupEnv("SNAPSHOT_NUMBER")
	if !validate {
		panic("You must set SNAPSHOT_NUMBER")
	}
	snapshot_number, err := strconv.Atoi(string_snapshot_number)
	if err != nil {
		panic(err)
	}

	export_to_s3, validate := os.LookupEnv("EXPORT_TO_S3")
	if !validate {
		panic("You must set EXPORT_TO_S3")
	}

	my_snapshot_config := snapshot_config{
		scw_access_key:  scw_access_key,
		scw_secret_key:  scw_secret_key,
		organizationID:  organizationID,
		project_id:      project_id,
		snapshot_region: snapshot_region,
		snapshot_az:     snapshot_az,
		disk_id:         disk_id,
		snapshot_number: snapshot_number,
		export_to_s3:    export_to_s3,
	}

	return my_snapshot_config
}

func priv_create_snapshot(my_snapshot_config snapshot_config, snapshot_name string, snapshot_tags *[]string, instanceApi *instance.API) (string, error) {
	// Create Snapshot
	fmt.Println(time.Now(), "- Creating snapshot")
	create_snapshot, err := instanceApi.CreateSnapshot(&instance.CreateSnapshotRequest{
		Zone:     my_snapshot_config.snapshot_az,
		Name:     snapshot_name,
		VolumeID: &my_snapshot_config.disk_id,
		Project:  &my_snapshot_config.project_id,
		Tags:     snapshot_tags,
	})
	if err == nil {
		// Print Snapshot Name and ID
		fmt.Println(time.Now(), "- Snapshot ID:", create_snapshot.Snapshot.ID)

		return create_snapshot.Snapshot.ID, err

	} else {
		return "", err
	}

}

func priv_wait_for_snapshot(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	// Wait for Snapshot
	fmt.Println(time.Now(), "- Waiting for snapshot")
	_, err := instanceApi.WaitForSnapshot(&instance.WaitForSnapshotRequest{
		SnapshotID: snapshot_id,
		Zone:       zone,
	})
	return err
}

func priv_export_to_s3(snapshot_id string, snapshot_name string, instanceApi *instance.API, zone scw.Zone) error {
	bucket, validate := os.LookupEnv("BUCKET_NAME")
	if !validate {
		err := errors.New("error, can't export snapshot to S3, you must set bucket_name")
		return err
	} else {
		// Export Snapshot to S3 Bucket
		fmt.Println(time.Now(), "- Exporting Snapshot to:", bucket)
		export_snapshot, err := instanceApi.ExportSnapshot(&instance.ExportSnapshotRequest{
			Zone:       zone,
			SnapshotID: snapshot_id,
			Bucket:     bucket,
			Key:        snapshot_name,
		})
		if err != nil {
			return err
		} else {
			fmt.Println(time.Now(), "- Export task ID:", export_snapshot.Task.ID)

			// Wait for export
			err = priv_wait_for_export(snapshot_id, instanceApi, zone)

			return err
		}
	}
}

func priv_wait_for_export(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	fmt.Println(time.Now(), "- Waiting for Export")
	get_snapshot, err := instanceApi.GetSnapshot(&instance.GetSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshot_id,
	})
	if err != nil {
		return err
	} else {

		fmt.Println(time.Now(), "- Waiting for Snapshot export")
		for get_snapshot.Snapshot.State != "available" {
			fmt.Println(time.Now(), "- Snapshot Status: ", get_snapshot.Snapshot.State)
			get_snapshot, err = instanceApi.GetSnapshot(&instance.GetSnapshotRequest{
				Zone:       zone,
				SnapshotID: snapshot_id,
			})
			if err != nil {
				return err
			}
			time.Sleep(30 * time.Second)
		}

		return err
	}
}

func list_snapshot(my_snapshot_config snapshot_config, instanceApi *instance.API) (*instance.ListSnapshotsResponse, error) {
	fmt.Println(time.Now(), "- Get Snapshot List for Disk:", my_snapshot_config.disk_id)

	list_snapshot, err := instanceApi.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone:         my_snapshot_config.snapshot_az,
		BaseVolumeID: &my_snapshot_config.disk_id,
	})

	return list_snapshot, err
}

func priv_delete_snapshot(snapshot_id string, instanceApi *instance.API, zone scw.Zone) error {
	fmt.Println(time.Now(), " - Deleting snapshot ", snapshot_id)
	err := instanceApi.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		Zone:       zone,
		SnapshotID: snapshot_id,
	})
	return err
}

func main() {

	my_snapshot_config := get_env_var()

	// Forge snapshot name
	snapshot_name := fmt.Sprint("snapshot.", time.Now().Unix())

	// tags array
	snapshot_tags := []string{"automatic"}

	// Allow us to skip task if a previous task failed
	// We always keep the cleaning task to avoid having too many snapshots
	abort_mission := false

	// Create a Scaleway client
	client, err := scw.NewClient(
		scw.WithDefaultOrganizationID(my_snapshot_config.organizationID),
		scw.WithAuth(my_snapshot_config.scw_access_key, my_snapshot_config.scw_secret_key),
		scw.WithDefaultRegion(scw.Region(my_snapshot_config.snapshot_region)),
	)
	if err != nil {
		panic(err)
	}

	// Create SDK objects for Scaleway Instance product
	instanceApi := instance.NewAPI(client)

	// Create Snapshot
	snapshot_id, err := priv_create_snapshot(my_snapshot_config, snapshot_name, &snapshot_tags, instanceApi)
	if err != nil {
		fmt.Println(err)
		abort_mission = true
	}

	// Wait for Snapshot creation
	if !abort_mission {
		err = priv_wait_for_snapshot(snapshot_id, instanceApi, my_snapshot_config.snapshot_az)
		if err != nil {
			fmt.Println(err)
			abort_mission = true
		}
	}

	if !abort_mission {
		// Export Snapshot to S3
		if my_snapshot_config.export_to_s3 == "true" {

			err = priv_export_to_s3(snapshot_id, snapshot_name, instanceApi, my_snapshot_config.snapshot_az)
			if err != nil {
				panic(err)
			}
		}
	}

	// LIST Snapshot
	snapshot_list, err := list_snapshot(my_snapshot_config, instanceApi)
	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now(), "- Number of Snapshot:", snapshot_list.TotalCount)

	// TODO Use an array to sort every snapshot
	oldest := snapshot_list.Snapshots[0]
	for _, snap := range snapshot_list.Snapshots {
		if snap.CreationDate.Unix() < oldest.CreationDate.Unix() {
			if slices.Contains(snap.Tags, snapshot_tags[0]) {
				oldest = snap
			}
		}
	}

	// TODO add the possibility to delete multiple snapshot if the number is higer than snapshot_number + 1
	// TODO use the tag to protect manually made snapshots
	if snapshot_list.TotalCount > uint32(my_snapshot_config.snapshot_number) {
		// DELETE Snapshot
		err = priv_delete_snapshot(oldest.ID, instanceApi, my_snapshot_config.snapshot_az)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("No Snapshot to delete")
	}
}
