module.exports = {
  printWidth: 100,
  singleQuote: true,
  overrides: [
    {
      files: '*.css',
      parser: 'css',
    },
    {
      files: '*.json',
      parser: 'json',
    },
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
