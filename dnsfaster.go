package main

// https://stackoverflow.com/questions/36417199/how-to-broadcast-message-using-channel

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "errors"
    "github.com/miekg/dns"
    "time"
    "math/rand"
    "strings"
)

const SEND_RESULTS = "SeNd ReSuLtS"
const WORKER_EXIT = "~"
const WORKER_NOTIFY_EXIT = "!~"
const SEPARATOR = "-------------------------------------------------------"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func getDNSList(fp string) ([]string, error) {
    file, err := os.Open(fp)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("[!] Can't open file: %s", fp))
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    return lines, nil
}

func workerResolverChecker(dc chan string, receiver chan *Result, base_domain string) {
    //var rtts []float64
    var domain string

    c := dns.Client{}
    m := dns.Msg{}

    for {
        resolver, ok := <-dc
        if !ok || resolver == WORKER_EXIT {
            break
        }

        if resolver == WORKER_NOTIFY_EXIT {
            receiver<-nil
            break
        }
        /*
        if resolver == SEND_RESULTS {
            for _, rtt := range rtts {
                if rtt != 0 {
                    ret<-rtt
                }
            }
            rtts = nil // delete the slice
            <-resume // wait to resume
            continue
        }*/

        domain = strings.Join([]string{RandStringBytes(5), ".", base_domain}, "")

        m.SetQuestion(domain + ".", dns.TypeA)
        _, rtt, err := c.Exchange(&m, resolver+":53")
        result := new(Result)
        result.dns = resolver
        if err == nil {
            result.rtt = float64(rtt/time.Microsecond)
            // TODO: send result to receiverService
            //rtts = append(rtts, float64(rtt/time.Microsecond))
        } else {
            result.rtt = -1
            //rtts = append(rtts, -1) // needed so all tests have a value
        }
        receiver<-result
    }

}

func printHeader(num_workers int, num_tests int, test_domain string, filepath string) {
    fmt.Println("Starting dnsfaster:")
    fmt.Println(SEPARATOR)
    fmt.Printf("| %d threads | %d tests | domain: %s | in file: %s |\n",
        num_workers, num_tests, test_domain, filepath)
    fmt.Println(SEPARATOR)
    fmt.Println("|              ip | avg micros | Rate |  Succ |  Fail | Action")
    fmt.Println(SEPARATOR)
}

type Result struct {
    dns string
    rtt float64
}

type ResultStats struct {
    dns string
    rtt float64
    succ int
    fail int
}

func receiverService(rcv chan *Result, done chan bool) {
    results := make(map[string]*ResultStats)

    for {
        result, ok := <-rcv
        if !ok || result == nil{
            break
        }

        _, prs := results[result.dns]
        if !prs {
            results[result.dns] = new(ResultStats)
        }

        cur := results[result.dns]

        if result.rtt == -1 {
            cur.fail++
            continue
        }

        cur.rtt += result.rtt
        cur.succ++
    }

    for _, r := range results {
        if r.rtt != 0 {
            r.rtt = r.rtt / float64(r.succ)
        }
        num_tests := r.succ + r.fail
        fmt.Printf("| %15s | %10v | %3d%% | %5d | %5d |\n", r.dns, int(r.rtt), r.succ*100/num_tests, r.succ, r.fail)
    }
    done<-true
}

func distributorService(num_workers int, num_tests int, test_domain string, filepath string) {

    rand.Seed(time.Now().UnixNano())

    printHeader(num_workers, num_tests, test_domain, filepath)

    resolvers, err := getDNSList(filepath)
    if err != nil {
        fmt.Println(err)
        return
    }

    //ret := make(chan float64)
    dc := make(chan string, 1000)
    //resume := make(chan bool)

    receiver := make(chan *Result, 100)
    rcvDone := make(chan bool)

    go receiverService(receiver, rcvDone)

    for i := 0; i < num_workers; i++ {
        go workerResolverChecker(dc, receiver, test_domain)
    }

    for _, dns := range resolvers {
        for i := 0; i < num_tests; i++ {
            dc<-dns
        }
        /*
        for i := 0; i < num_workers; i++ {
            dc<-SEND_RESULTS
        }

        var j int
        var avg float64
        for i := 0; i < num_tests; i++ {
            tmp := <-ret
            if tmp != -1 {
                avg += tmp
                j++
            }
        }

        for i := 0; i < num_workers; i++ {
            resume<-true
        }
        
        if (avg != 0) {
            avg = avg / float64(j)
        }

        fmt.Printf("| %15s | %10v | %3d%% | %5d | %5d |\n", dns, int(avg), j*100/num_tests, j, num_tests - j)
        avg = 0
        */
    }

    for i := 0; i < num_workers; i++ {
        if i+1 == num_workers { // last worker notifies receiver
            dc<-WORKER_NOTIFY_EXIT
            continue
        }
        dc<-WORKER_EXIT
    }

    <-rcvDone
}

func main() {
    if len(os.Args) < 5 {
        fmt.Println("usage: ./dnsfaster <input filepath> <num_workers> <num_tests> <test domain>")
        return
    }
    filepath := os.Args[1]

    num_workers, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Println("Invalid num workers :", os.Args[2])
        return
    }

    num_tests, err := strconv.Atoi(os.Args[3])
    if err != nil {
        fmt.Println("Invalid num tests :", os.Args[3])
        return
    }

    domain := os.Args[4]

    distributorService(num_workers, num_tests, domain, filepath)
}
