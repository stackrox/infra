import json
import sys

from tabulate import tabulate

DAILY_COST_MAP = {
    "demo": 33,
    "gke-default": 10,
    "openshift-4": 29,
    "openshift-4-demo": 53,
    "openshift-4-perf-scale": 70,
    "openshift-multi": 10,
    "osd-on-gcp": 35,
    "qa-demo": 33,
    "osd-on-aws": 50,
    "rosa": 97,
    "eks": 13,
    "aks": 17,
    "aro": 53,
}

def read_usage_from_stdin():
    content = ""
    print("hello")
    for line in sys.stdin:
        print(1, line)
        content += line

    print(content)

    return json.loads(content)

def calculate_costs(usage):
    costs = []
    for x in usage:
        current = {
            "environment": x["environment"],
            "flavor": x["flavor"],
            "total usage (days)": x["total_days_consumed"],
        }

        current["cost (usd)"] = float(x["total_days_consumed"]) * DAILY_COST_MAP[x["flavor"]]
        costs.append(current)

    return costs

def main():
    usage = read_usage_from_stdin()
    cost_per_flavor_env = calculate_costs(usage)

    print(tabulate(
        cost_per_flavor_env,
        headers="keys",
        tablefmt="simple_outline",
        floatfmt=".2f"
    ))


if __name__ == "__main__":
    main()
