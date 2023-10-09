import json
import sys

from tabulate import tabulate

# Estimation based on flavor defaults
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
    for line in sys.stdin:
        content += line

    try:
        return json.loads(content)
    except Exception as e:
        raise("an exception occured while reading stdin:", e)

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

    cost_per_flavor_env.sort(key=lambda x: x["flavor"])

    print("```")
    print(tabulate(
        cost_per_flavor_env,
        headers="keys",
        tablefmt="github",
        floatfmt=".2f"
    ))
    print("```")


if __name__ == "__main__":
    main()
