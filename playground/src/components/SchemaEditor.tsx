"use client";

import dynamic from "next/dynamic";
import { SchemaType } from "@/lib/types";
import { MONACO_LANGUAGE } from "@/lib/samples";

const Editor = dynamic(() => import("@monaco-editor/react"), { ssr: false });

interface Props {
  label: string;
  subtitle: string;
  value: string;
  onChange: (value: string) => void;
  schemaType: SchemaType;
}

export default function SchemaEditor({ label, subtitle, value, onChange, schemaType }: Props) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-baseline gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">
          {label}
        </span>
        <span className="text-xs text-slate-500">{subtitle}</span>
      </div>
      <div className="rounded-lg overflow-hidden border border-slate-700 h-[400px]">
        <Editor
          height="100%"
          language={MONACO_LANGUAGE[schemaType]}
          value={value}
          onChange={(v) => onChange(v ?? "")}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            fontSize: 13,
            lineHeight: 20,
            fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
            scrollBeyondLastLine: false,
            wordWrap: "off",
            automaticLayout: true,
            tabSize: 2,
          }}
        />
      </div>
    </div>
  );
}
