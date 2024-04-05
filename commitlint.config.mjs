export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'scope-enum': [
      2,
      'always',
      [
        'operator',
        'agent',
        'vault-api',
        'deps'
      ]
    ],
    'signed-off-by': [
      2,
      'always',
      'Signed-off-by:'
    ],
    'trailer-exists': [
      2,
      'always',
      'Signed-off-by:'
    ],
    'header-max-length': [
      2,
      'always',
      180
    ]
  }
};
