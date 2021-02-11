// eslint-disable-next-line import/prefer-default-export
export function getOneClickFlavor(): {
  ID: string;
  Name: string;
  Description: string;
  Availability: string;
  Parameters: {
    'k8s-version': {
      Name: string;
      Description: string;
      Value: string;
      Optional: boolean;
      Order: number;
    };
    'main-image': {
      Name: string;
      Description: string;
      Value: string;
      Order: number;
      Help: string;
      Optional: boolean;
    };
    name: { Name: string; Description: string; Value: string; Order: number };
    'scanner-db-image': {
      Name: string;
      Description: string;
      Optional: boolean;
      Order: number;
      Help: string;
    };
    'scanner-image': {
      Name: string;
      Description: string;
      Optional: boolean;
      Order: number;
      Help: string;
    };
  };
  Artifacts: {
    kubeconfig: { Name: string; Description: string };
    roxctl: { Name: string; Tags: { internal: unknown } };
    tfstate: { Name: string; Description: string };
    url: { Name: string; Description: string; Tags: { url: unknown } };
  };
} {
  return {
    ID: 'one-click-release-demo',
    Name: 'One-Click Release Demo',
    Description: 'Demo running the StackRox version in the latest release tag',
    Availability: 'stable',
    Parameters: {
      'k8s-version': {
        Name: 'k8s-version',
        Description: 'kubernetes version',
        Value: 'default',
        Optional: true,
        Order: 5,
      },
      'main-image': {
        Name: 'main-image',
        Description: 'StackRox Central image Docker name',
        Value: 'docker.io/stackrox/main:3.0.55.0-rc.7',
        Order: 2,
        Optional: true,
        Help: '',
      },
      name: {
        Name: 'name',
        Description: 'cluster name',
        Value: 'example1',
        Order: 1,
      },
      'scanner-db-image': {
        Name: 'scanner-db-image',
        Description: 'StackRox Scanner DB image Docker name',
        Optional: true,
        Order: 4,
        Help: 'If unspecified, this will be derived from the central image',
      },
      'scanner-image': {
        Name: 'scanner-image',
        Description: 'StackRox Scanner image Docker name',
        Optional: true,
        Order: 3,
        Help: 'If unspecified, this will be derived from the central image',
      },
    },
    Artifacts: {
      kubeconfig: {
        Name: 'kubeconfig',
        Description: 'Kube config for connecting to cluster',
      },
      roxctl: {
        Name: 'roxctl',
        Tags: {
          internal: {},
        },
      },
      tfstate: {
        Name: 'tfstate',
        Description: 'Terraform state file',
      },
      url: {
        Name: 'url',
        Description: 'URL of StackRox UI',
        Tags: {
          url: {},
        },
      },
    },
  };
}
