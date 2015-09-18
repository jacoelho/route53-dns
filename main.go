package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/codegangsta/cli"
)

type updateConfig struct {
	domain string
	name   string
}

func updateDnsRecord(config *updateConfig) error {
	svcMetadata := ec2metadata.New(&ec2metadata.Config{})

	instanceIp, err := svcMetadata.GetMetadata("local-ipv4")
	if err != nil {
		log.Fatal(err)
	}

	region, err := svcMetadata.Region()
	if err != nil {
		log.Fatal(err)
	}

	svc := route53.New(
		&aws.Config{
			Region: aws.String(region),
		},
	)

	if config.domain[len(config.domain)-1] != '.' {
		config.domain = fmt.Sprintf("%s.", config.domain)
	}

	listDomains, err := svc.ListHostedZonesByName(
		&route53.ListHostedZonesByNameInput{
			DNSName: aws.String(config.domain),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	if len(listDomains.HostedZones) != 1 {
		log.Fatal("invalid domain")
	}

	tmp := strings.Split(*listDomains.HostedZones[0].Id, "/")
	domainId := tmp[len(tmp)-1]

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(config.name),
						Type: aws.String("A"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(instanceIp),
							},
						},
						TTL: aws.Int64(600),
					},
				},
			},
			Comment: aws.String("ResourceDescription"),
		},
		HostedZoneId: aws.String(domainId),
	}
	_, err = svc.ChangeResourceRecordSets(params)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "route53-dns"
	app.Usage = "update route53 dns record"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "dnsname",
			Usage: "dns name to update/create",
		},
		cli.StringFlag{
			Name:  "domain",
			Usage: "record domain",
		},
	}

	app.Action = func(c *cli.Context) {
		domain := c.String("domain")
		dnsName := c.String("dnsname")

		if len(domain) == 0 {
			log.Fatal("invalid domain name")
		}

		if len(dnsName) == 0 {
			log.Fatal("invalid dns name")
		}

		err := updateDnsRecord(&updateConfig{
			name:   dnsName,
			domain: domain,
		})
		if err != nil {
			log.Fatal("failed to update")
		}
	}

	app.Run(os.Args)
}
