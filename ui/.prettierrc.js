module.exports = {
  printWidth: 100,
  singleQuote: true,
  overrides: [
    {
      files: '*.md',
      parser: 'markdown',
      options: {
        printWidth: 80,
        proseWrap: 'always',
      },
    },
  ],
};
