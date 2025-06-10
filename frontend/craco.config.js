module.exports = {
  webpack: {
    configure: (webpackConfig, { env, paths }) => {
      // Enable hot module replacement
      if (env === 'development') {
        webpackConfig.devServer = {
          ...webpackConfig.devServer,
          hot: true,
          liveReload: true,
          watchFiles: {
            paths: ['src/**/*', 'public/**/*'],
            options: {
              usePolling: false,
              interval: 1000,
            },
          },
        };
      }
      return webpackConfig;
    },
  },
  devServer: {
    hot: true,
    liveReload: true,
    watchFiles: ['src/**/*', 'public/**/*'],
    client: {
      overlay: {
        errors: true,
        warnings: false,
      },
    },
  },
}; 