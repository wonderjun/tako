/// <reference types="vite/client" />

// vite 8 的 client.d.ts 不再内建 ImportMeta.env 声明，项目侧自行补充标准字段。
interface ImportMetaEnv {
  readonly DEV: boolean;
  readonly MODE: string;
  readonly PROD: boolean;
  readonly SSR: boolean;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
