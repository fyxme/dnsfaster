package main

import (
    "bufio"
    "fmt"
    "os"
    "errors"
    "github.com/miekg/dns"
    "time"
    "math/rand"
    "strings"
    "flag"
)

const WORKER_EXIT = "~"
const WORKER_NOTIFY_EXIT = "!~"

const SEPARATOR = "-------------------------------------------------------"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

type TestInfo struct {
    domain string
    dns string
    rtt float64
}

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

func printHeader(num_workers int, num_tests int, test_domain string, filepath string) {
    fmt.Println(`
           _            __          _
          | |          / _|        | |
        __| |_ __  ___| |_ __ _ ___| |_ ___ _ __
       / _' | '_ \/ __|  _/ _' / __| __/ _ \ '__|
      | (_| | | | \__ \ || (_| \__ \ ||  __/ |
       \__,_|_| |_|___/_| \__,_|___/\__\___|_|

    `)

    fmt.Println(SEPARATOR)
    fmt.Printf("| %7d threads | domain  : %23s |\n", num_workers, test_domain)
    fmt.Printf("| %7d tests   | in file : %23s |\n", num_tests, filepath)
    fmt.Println(SEPARATOR)
    fmt.Println("|              ip | avg micros | Rate |  Succ |  Fail |")
    fmt.Println(SEPARATOR)
}

func workerResolverChecker(dc chan *TestInfo, receiver chan *TestInfo, base_domain string) {
    for {
        test, ok := <-dc
        if !ok || test.dns == WORKER_EXIT {
            break
        }

        if test.dns == WORKER_NOTIFY_EXIT {
            receiver<-nil
            break
        }

        c := dns.Client{}
        m := dns.Msg{}

        m.SetQuestion(test.domain + ".", dns.TypeA)
        r, rtt, err := c.Exchange(&m, test.dns + ":53")

        // make sure the server responds and returns no entry
        if err == nil && r != nil && r.Rcode == dns.RcodeNameError {
            test.rtt = float64(rtt/time.Microsecond)
        }
        receiver<-test
    }
}


func receiverService(rcv chan *TestInfo, done chan bool, num_tests int, outfp string) {
    results := make(map[string]*ResultStats)

    defer func() { done<-true }() // close the channel once done

    file, err := os.OpenFile(outfp, os.O_WRONLY | os.O_CREATE, 0666)
    if err != nil {
        fmt.Println("[!] Can't open file: ", outfp)
        return
    }

    defer file.Close()

    w := bufio.NewWriter(file)

    if _, err := w.WriteString("ip,avg rtt,success rate,success,failure\n"); err != nil {
        panic(err)
    }

    for {
        result, ok := <-rcv
        if !ok || result == nil{
            break
        }

        _, prs := results[result.dns]
        if !prs {
            results[result.dns] = new(ResultStats)
            results[result.dns].dns = result.dns
        }

        cur := results[result.dns]

        if result.rtt == -1 {
            cur.fail++
        } else {
            cur.rtt += result.rtt
            cur.succ++
        }

        if cur.succ + cur.fail == num_tests {
            if cur.rtt != 0 {
                cur.rtt = cur.rtt / float64(cur.succ)
            }
            succ_p := cur.succ*100/num_tests
            fmt.Printf("| %15s | %10v | %3d%% | %5d | %5d |\n", cur.dns, int(cur.rtt), succ_p, cur.succ, cur.fail)

            if succ_p >= 95 { // only keeps above or equal to 95% success rate
                s := fmt.Sprintf("%s,%v,%d,%d,%d\n", cur.dns, int(cur.rtt), succ_p, cur.succ, cur.fail)
                if _, err := w.WriteString(s); err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }
            }
        }
    }
    if err := w.Flush(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fmt.Println(SEPARATOR)
}

func distributorService(num_workers int, num_tests int, test_domain string, infp string, outfp string) {

    rand.Seed(time.Now().UnixNano())

    printHeader(num_workers, num_tests, test_domain, infp)

    resolvers, err := getDNSList(infp)
    if err != nil {
        fmt.Println(err)
        return
    }

    // pregenerate test cases
    var domains []string
    for i := 0; i < num_tests; i++ {
        domains = append(domains, strings.Join([]string{RandStringBytes(8), ".", test_domain}, ""))
    }

    dc := make(chan *TestInfo, 1000)
    receiver := make(chan *TestInfo, 250)

    rcvDone := make(chan bool)

    go receiverService(receiver, rcvDone, num_tests, outfp)

    for i := 0; i < num_workers; i++ {
        go workerResolverChecker(dc, receiver, test_domain)
    }

    for i := 0; i < num_tests; i++ {
        for _, dns := range resolvers {
            test := new(TestInfo)
            test.dns = dns
            test.domain = domains[i]
            test.rtt = -1
            dc<-test
        }
    }

    for i := 0; i < num_workers; i++ {
        test := new(TestInfo)
        if i+1 == num_workers { // last worker notifies receiver
            test.dns = WORKER_NOTIFY_EXIT
        } else {
            test.dns = WORKER_EXIT
        }
        dc<-test
    }

    <-rcvDone
}

func main() {
    infp := flag.String("in", "", "The input filepath")
    outfp  := flag.String("out", "", "The output file")
    num_workers := flag.Int("workers", 10, "Number of workers")
    num_tests := flag.Int("test", 100, "Number of tests per dns server")
    domain := flag.String("domain", "example.com", "Domain name to test against")
    flag.Parse();

    // no need to test input and output file as https://stackoverflow.com/a/22483001 suggests
    if *num_workers < 1 {
        fmt.Fprintf(os.Stderr, "[!] Invalid number of workers: %d\n", *num_workers)
        return
    }

    if *num_tests < 1 {
        fmt.Fprintf(os.Stderr, "[!] Invalid number of tests: %d\n", *num_tests)
        return
    }

    distributorService(*num_workers, *num_tests, *domain, *infp, *outfp)
}
