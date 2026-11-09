import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';

export default defineConfig({
  plugins: [pluginReact()],
  // source: {
  //   entry: {
  //     index: './src/index.tsx',
  //     file: './src/file.tsx',
  //   }
  // },
  html: {
    title: "Garbage In; Garbage Out"
  },

  // Switch the output path because I'm too lazy to setup a proper build for it.
  output: {
    distPath: {
      root: "../dist",
    },
    cleanDistPath: {
      enable: true
    },
    sourceMap:
    {
      js: 'source-map'
    },
  }
});
