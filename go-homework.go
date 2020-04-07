package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.etcd.io/etcd/clientv3"
)

const ETCD_KEY_NAME = "operation"

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/calc/sum/{valA}/{valB}", sum)
	router.HandleFunc("/calc/subtract/{valA}/{valB}", subtract)
	router.HandleFunc("/calc/multiply/{valA}/{valB}", multiply)
	router.HandleFunc("/calc/divide/{valA}/{valB}", divide)
	router.HandleFunc("/calc/history", history)
	router.HandleFunc("/calc/deleteAllUserData", deleteUserData)

	http.ListenAndServe(":8080", router)
}

func deleteUserData(w http.ResponseWriter, r *http.Request) {
	cli, ctx, cancel, err := etcdConnect(w)
	if err != nil {
		fmt.Fprintf(w, "Error opening ETCD connection\n")
		closeEtcd(w, cli, cancel)
		return
	}

	result, getErr := listHistoryOperations(cli, ctx, buildOpts())
	if getErr != nil {
		fmt.Fprintf(w, "Error listing operations\n")
		closeEtcd(w, cli, cancel)
		return
	}

	if len(result.Kvs) <= 0 {
		fmt.Fprintf(w, "Nothing to delete here\n")
		closeEtcd(w, cli, cancel)
		return
	}

	for _, item := range result.Kvs {
		_, delErr := cli.Delete(ctx, string(item.Key))
		if delErr != nil {
			fmt.Fprintf(w, "Error deleting keys\n")
		} else {
			fmt.Fprintf(w, "Key deleted successfully: %s\n", string(item.Key))
		}
	}

	closeEtcd(w, cli, cancel)
}

func buildOpts() []clientv3.OpOption {
	return []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	}
}

func history(w http.ResponseWriter, r *http.Request) {
	cli, ctx, cancel, err := etcdConnect(w)
	if err != nil {
		fmt.Fprintf(w, "Error opening ETCD connection\n")
		closeEtcd(w, cli, cancel)
		return
	}

	result, getErr := listHistoryOperations(cli, ctx, buildOpts())
	if getErr != nil {
		fmt.Fprintf(w, "Error listing operations\n")
		closeEtcd(w, cli, cancel)
		return
	}
	for _, item := range result.Kvs {
		fmt.Fprintf(w, "Stored key / value: %s = %s\n", string(item.Key), string(item.Value))
	}

	closeEtcd(w, cli, cancel)
}

func listHistoryOperations(cli *clientv3.Client, ctx context.Context, opts []clientv3.OpOption) (*clientv3.GetResponse, error) {
	return cli.Get(ctx, ETCD_KEY_NAME, opts...)
}

func putValueEtcd(w http.ResponseWriter, value int64) {
	cli, ctx, cancel, err := etcdConnect(w)
	if err != nil {
		fmt.Fprintf(w, "Error opening ETCD connection\n")
		closeEtcd(w, cli, cancel)
		return
	}

	keyToEtcd := fmt.Sprintf(ETCD_KEY_NAME+"_%02d", rand.Intn(1000))
	fmt.Fprintf(w, "Key to ETCD: %s\n", string(keyToEtcd))
	fmt.Fprintf(w, "Value to ETCD: %d\n", value)

	resPut, errPut := cli.Put(ctx, keyToEtcd, strconv.FormatInt(value, 10))
	if errPut != nil {
		fmt.Fprintf(w, "Error putting operation on ETCD: %s\n", errPut)
	} else {
		fmt.Fprintf(w, "Successfully PUT on ETCD: %s \n", resPut.OpResponse().Put())
	}

	closeEtcd(w, cli, cancel)
}

func etcdConnect(w http.ResponseWriter) (*clientv3.Client, context.Context, context.CancelFunc, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379", "etcd:22379", "etcd:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintf(w, "Error starting connection with ETCD: %s\n", err)
	} else {
		fmt.Fprintf(w, "ETCD Client connected\n")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	return cli, ctx, cancel, err
}

func closeEtcd(w http.ResponseWriter, cli *clientv3.Client, cancel context.CancelFunc) {
	cancel()
	errClose := cli.Close()
	if errClose != nil {
		fmt.Fprintf(w, "Error closing ETCD client\n")
	} else {
		fmt.Fprintf(w, "ETCD Client disconnected\n")
	}
}

func subtract(w http.ResponseWriter, r *http.Request) {
	intValA, intValB, err := parseAndCheckValues(w, r)
	if err != nil {
		return
	}

	result := intValA - intValB
	putValueEtcd(w, result)

	fmt.Fprintf(w, "The subtraction result is: %d\n", result)
}

func multiply(w http.ResponseWriter, r *http.Request) {
	intValA, intValB, err := parseAndCheckValues(w, r)
	if err != nil {
		return
	}

	result := intValA * intValB
	putValueEtcd(w, result)

	fmt.Fprintf(w, "The multiplication result is: %d\n", result)
}

func divide(w http.ResponseWriter, r *http.Request) {
	intValA, intValB, err := parseAndCheckValues(w, r)
	if err != nil {
		return
	}

	result := intValA / intValB
	putValueEtcd(w, result)

	fmt.Fprintf(w, "The division result is: %d\n", result)
}

func sum(w http.ResponseWriter, r *http.Request) {
	intValA, intValB, err := parseAndCheckValues(w, r)
	if err != nil {
		return
	}

	result := intValA + intValB
	putValueEtcd(w, result)

	fmt.Fprintf(w, "The sum result is: %d\n", result)
}

func parseAndCheckValues(w http.ResponseWriter, r *http.Request) (int64, int64, error) {
	vars := mux.Vars(r)

	intValA, err := parseAndCheckIntValue(vars["valA"], w)
	if err != nil {
		return 0, 0, err
	}
	intValB, err := parseAndCheckIntValue(vars["valB"], w)
	if err != nil {
		return 0, 0, err
	}
	return intValA, intValB, nil
}

func parseAndCheckIntValue(val string, w http.ResponseWriter) (int64, error) {
	varA, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "Parameter must '%s' be an Integer\n", val)
		return 0, err
	}
	return varA, err
}
