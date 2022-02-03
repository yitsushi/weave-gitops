const versions = require("./versions.json");
/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: "Weave GitOps",
  tagline: "Weave GitOps Documentation",
  url: "https://docs.gitops.weave.works/",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon_150px.png",
  organizationName: "weaveworks", // Usually your GitHub org/user name.
  projectName: "weave-gitops-docs", // Usually your repo name.
  plugins: [
    () => ({
      // Load yaml files as blobs
      configureWebpack: function () {
        return {
          module: {
            rules: [
              {
                test: /\.yaml$/,
                use: [
                  {
                    loader: "file-loader",
                    options: { name: "assets/files/[name]-[hash].[ext]" },
                  },
                ],
              },
            ],
          },
        };
      },
    }),
  ],
  themeConfig: {
    navbar: {
      title: "Weave GitOps",
      logo: {
        alt: "Weave GitOps Logo",
        src: "img/weave-logo.png",
      },
      items: [
        {
          type: "doc",
          docId: "intro",
          position: "left",
          label: "Introduction",
        },
        {
          type: "doc",
          docId: "installation",
          position: "left",
          label: "Installation",
        },
        {
          type: "doc",
          docId: "getting-started",
          position: "left",
          label: "Getting Started",
        },
        {
          type: "doc",
          docId: "aws-marketplace",
          position: "left",
          label: "AWS Marketplace",
        },
        {
          type: "docsVersionDropdown",
          position: "right",
          dropdownActiveClassDisabled: true,
        },
        {
          href: "https://github.com/weaveworks/weave-gitops",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Support",
          items: [
            {
              label: "Contact Us",
              href: "mailto:support@weave.works",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Weaveworks`,
    },
    algolia: {
      apiKey: process.env.ALGOLIA_API_KEY,
      indexName: "weave",
      // Needed to handle the different versions of docs
      contextualSearch: true,

      // Optional: Algolia search parameters
      // searchParameters: {
      //   facetFilters: ['type:content']
      // },
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          editUrl: "https://github.com/weaveworks/weave-gitops-docs/edit/main/",
          lastVersion: versions[0],
          versions: {
            current: {
              label: "main",
            },
          },
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            "https://github.com/weaveworks/weave-gitops-docs/edit/main/blog/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        gtag: {
          // You can also use your "G-" Measurement ID here.
          // Bogus commit to trigger a build
          trackingID: process.env.GA_KEY,
          // Optional fields.
          anonymizeIP: true, // Should IPs be anonymized?
        },
      },
    ],
  ],
};
