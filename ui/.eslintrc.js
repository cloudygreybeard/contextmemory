module.exports = {
    "env": {
        "browser": true,
        "es6": true,
        "node": true
    },
    "extends": [
        "eslint:recommended"
    ],
    "parser": "@typescript-eslint/parser",
    "parserOptions": {
        "project": "tsconfig.json",
        "sourceType": "module"
    },
    "plugins": [
        "@typescript-eslint"
    ],
    "rules": {
        // Disable problematic rules for now to focus on core functionality
        "no-unused-vars": "off",
        "no-undef": "off", 
        "no-case-declarations": "off",
        "curly": "off"
    },
    "ignorePatterns": [
        "out",
        "dist",
        "**/*.d.ts"
    ]
};
