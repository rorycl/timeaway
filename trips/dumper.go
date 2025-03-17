package trips

//
// import (
// 	"log"
// 	"os"
//
// 	"github.com/sanity-io/litter"
// )
//
// // dumper dumps trips to a file.
// func dumper(trips *Trips) {
// 	f, err := os.Create("trips_dump.txt")
// 	if err != nil {
// 		log.Printf("dump file open error %v", err)
// 		return
// 	}
// 	defer f.Close()
// 	litter.Config.FormatTime = true
// 	litter.Config.DisablePointerReplacement = true
// 	tDump := litter.Sdump(trips)
// 	_, err = f.Write([]byte(tDump))
// 	if err != nil {
// 		log.Printf("dump write error %v", err)
// 		return
// 	}
// }
