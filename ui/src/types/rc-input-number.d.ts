declare module 'rc-input-number' {
  export type RcInputNumberProps = {
    name?: string;
    value?: number;
    min?: number;
    max?: number;
    step?: number | string;
    precision?: number;
    style?: object;
    readOnly?: boolean;
    disabled?: boolean;
    onChange?: (value: number) => void;
    'aria-label'?: string;
  };
  function NumericInput(props: RcInputNumberProps): import('react').ReactElement;
  export = NumericInput;
}
