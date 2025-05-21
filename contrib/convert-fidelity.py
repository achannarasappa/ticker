#!/usr/bin/env python3
import argparse
import csv
import decimal
import yaml
import logging
from decimal import Decimal, InvalidOperation

logger = logging.getLogger(__name__)

parser = argparse.ArgumentParser(
    prog="convert-fidelity.py",
    description="Converts a CSV file provided by Fidelity to the format used by https://github.com/achannarasappa/ticker",
)

parser.add_argument("filename", help="The name of the CSV file provided by fidelity")
args = parser.parse_args()

with open(args.filename, "r") as csvfile:
    c = csv.DictReader(csvfile)
    ticker_config = {
        "show-summary": True,
        "show-fundamentals": True,
        "show-tags": True,
        "show-separator": True,
        "show-holdings": True,
        "currency-summary-only": True,
        "interval": 10,
        "currency": "USD",
    }

    watchlist = []
    lots = []

    for itm in c:
        watchlist.append(itm["Symbol"])
        try:
            itmdict = {
                "symbol": itm["Symbol"],
                "quantity": float(itm["Quantity"]),
                "unit_cost": float(itm["Average Cost Basis"][1:]),
            }
            lots.append(itmdict)
        except ValueError:
            logger.warn(
                "Unable to convert average cost basis '%s' or quantity '%s' for %s"
                % (itm["Average Cost Basis"][1:], itm["Quantity"], itm["Symbol"])
            )

ticker_config.update({"watchlist": watchlist, "lots": lots})
print(yaml.dump(ticker_config))
