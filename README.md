This is a fork of terraform-provider-packet, which was reworked into [terraform-provider-equinix](https://github.com/equinix/terraform-provider-equinix) after Packet was acquired by Equinix and rebranded as Equinix Metal. The provider code was copied, unfortunately effectively removing record (git history) of my ([t0mk](https://github.com/t0mk/)'s) contributions (which I'm fine with, no grudge there). 

This repo is just a copy of the old Packet provider (by 2023 deprecated and in future probably removed), so that I can prove the open-source work that I've done on terraform-provider-packet. I was the largest contributor to the project between 2017 and 2021.

# Equinix Metal Terraform Provider

[![End of Life](https://img.shields.io/badge/Stability-EndOfLife-black.svg)](end-of-life-statement.md#end-of-life-statements)
[![GitHub release](https://img.shields.io/github/release/packethost/terraform-provider-packet/all.svg?style=flat-square)](https://github.com/packethost/terraform-provider-packet/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/packethost/terraform-provider-packet)](https://goreportcard.com/report/github.com/packethost/terraform-provider-packet)

[![Slack](https://slack.equinixmetal.com/badge.svg)](https://slack.equinixmetal.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/equinixmetal.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=equinixmetal)

<img src="https://metal.equinix.com/metal/images/logo/equinix-metal-full.svg" width="600px">

>### This repository is [End of Life](https://github.com/equinix-labs/equinix-labs/blob/master/end-of-life-statement.md) meaning that this software is no longer supported nor maintained by Equinix Metal or its community.



[Please review the Packet to Equinix provider migration guide](https://registry.terraform.io/providers/equinix/equinix/latest/docs/guides/migration_guide_packet).

The [Equinix provider](https://registry.terraform.io/providers/equinix/equinix/latest/docs) has full support for existing Terraform managed Metal ([Packet is now Equinix Metal](https://blog.equinix.com/blog/2020/10/06/equinix-metal-metal-and-more/)) resources once Terraform configuration and state are adapted.  The Equinix provider manages resources including Network Edge and Fabric in addition to Metal.
