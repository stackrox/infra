import re
import sys


FLAVOR_FILE="chart/infra-server/static/flavors.yaml"
LAST_VERSION_FILE="ACS_DEMO_VERSION"


class Version:
    def __init__(self, tag):
        pattern = re.compile("^(\d+\.)?(\d+\.)?(\*|\d+)$")
        if not pattern.match(tag):
            raise ValueError("invalid version format")

        parts = tag.split(".")
        self.major, self.minor, self.patch = int(parts[0]), int(parts[1]), int(parts[2])

    def __gt__(self, other):
        return (self.major, self.minor, self.patch) > (other.major, other.minor, other.patch)

    def __str__(self):
        return f"{self.major}.{self.minor}.{self.patch}"


def read_supported_demo_version():
    with open(LAST_VERSION_FILE) as f:
        return f.read()


def determine_change_required(last_version, new_version):
    if new_version > last_version:
        return True
    return False


def update_demo_versions(last_version, new_version):
    with open(FLAVOR_FILE, "r+") as f:
        content = f.read()
        content = content.replace(str(last_version), str(new_version))
        f.seek(0)
        f.write(content)
        f.truncate()

    with open(LAST_VERSION_FILE, "w") as f:
        f.write(f"{new_version}\n")


def run(new_version):
    last_version = Version(read_supported_demo_version())
    if determine_change_required(last_version, new_version):
        update_demo_versions(last_version, new_version)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.stderr.write("usage: update_demo_flavors.py <NEW_VERSION>\n")
        sys.exit(1)
    try:
        new_version = Version(sys.argv[1])
    except ValueError as e:
        sys.stderr.write(f"could not parse NEW_VERSION argument '{sys.argv[1]}': {str(e)}\n")
        sys.exit()

    run(new_version)
