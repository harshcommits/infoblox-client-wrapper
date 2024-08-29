package main

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	dc "dnsclient-poc/raw"
	rec "dnsclient-poc/raw/records"
)

func main() {

	// set up context as per gardener extension
	ctx := signals.SetupSignalHandler()

	client, err := dc.NewDNSClient(ctx, "admin", "btprpc_infoblox", "10.16.198.17") // older installation
	// client, err := dc.NewDNSClient(ctx, "admin", "infoblox", "10.16.198.191") // newer installation
	if err != nil {
		fmt.Println(err)
	}

	// values := []string{"12.11.100.9"}

	// err2 := client.CreateOrUpdateRecordSet(ctx, "default", "txt3.infobloxbtprpc", "infobloxbtprpc", rec.Type_TXT, values, 0)
	// if err2 != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("Record created successfully")
	// }

	del_rec := client.DeleteRecordSet(ctx, "infobloxbtprpc", "reca.infobloxbtprpc", rec.Type_A)
	if del_rec != nil {
		fmt.Println(del_rec)
	} else {
		fmt.Println("Deletion successful")
	}

	// records, err := client.GetRecordSet(false, rec.Type_A, "infobloxbtprpc")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// for _, rec := range records {
	// 	fmt.Println(rec)
	// }

	// zones := make(map[string]string)

	// zones, err = client.GetManagedZones(ctx)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// for _, zone := range zones {
	// 	fmt.Println(zone)
	// }

}
