import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

export default defineConfig({
  plugins: [vue()],
  // Configure build output
  build: {
    // Output directory for build assets relative to project root
    outDir: 'dist',
    // Specify the entry point for the build.
    // This tells Vite to bundle src/main.js and its dependencies.
    rollupOptions: {
      input: {
          main: path.resolve(__dirname, 'src/main.js'),
	  //main: path.resolve(__dirname, 'src/edit_story.js'),
      },
      output: {
        // Force a specific name for the main JS bundle to avoid hashes
        entryFileNames: `js/app.js`,
        // You can still use hashes for other chunks/assets if needed
        chunkFileNames: `js/[name]-[hash].js`,
        assetFileNames: (assetInfo) => {
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            return 'css/[name]-[hash][extname]'; // Example for other CSS files
          }
          return 'assets/[name]-[hash][extname]'; // Default for other assets
        },
      },
    },
    // Minify output for production builds.
    //minify: true,
    // Disable CSS code splitting for a single CSS file (Tailwind handles CSS).
    cssCodeSplit: false,
  },
  // Base path for assets when served in production
  base: '/static/',
});
