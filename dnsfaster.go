package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "errors"
    "github.com/miekg/dns"
    "time"
    "math/rand"
)

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

func workerResolverChecker(dc chan string, ret chan float64, resolver string) {
    var rtts []float64

    for {
        domain, ok := <-dc
        if !ok {
            break
        }
        c := dns.Client{}
        m := dns.Msg{}

        m.SetQuestion(domain + ".", dns.TypeA)
        _, rtt, err := c.Exchange(&m, resolver+":53")
        if err == nil {
            rtts = append(rtts, float64(rtt/time.Nanosecond))
        } else {
            rtts = append(rtts, -1) // needed so all tests have a value
        }
    }

    for _, rtt := range rtts {
        if rtt != 0 {
            ret<-rtt
        }
    }
}


func distributorService(num_workers int, num_tests int, test_domain string, filepath string) {

    rand.Seed(time.Now().UnixNano())

    fmt.Printf("Starting dnsfaster:\n| %d threads | %d tests | test domain: %s | input file: %s |", 
        num_workers, num_tests, test_domain, filepath)

    resolvers, err := getDNSList(filepath)
    if err != nil {
        fmt.Println(err)
        return
    }

    ret := make(chan float64)
    for _, dns := range resolvers {
        dc := make(chan string)
        var avg float64
        for i := 0; i < num_workers; i++ {
            go workerResolverChecker(dc, ret, dns)
        }
        for i := 0; i < num_tests; i++ {
            dc<-fmt.Sprintf("%s.%s", RandStringBytes(5), test_domain)
        }

        close(dc)
        var j int
        for i := 0; i < num_tests; i++ {
            tmp := <-ret
            if tmp != -1 {
                avg += tmp
                j++
            }
        }
        avg = avg / float64(j)
        fmt.Println("Final: ", dns, avg)
        fmt.Println("Test stats: [%d%%] %d Returned, %d Failed\n", j*100/num_tests, j, num_tests - j)
    }
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
