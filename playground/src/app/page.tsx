"use client";

import { useState } from "react";
import SchemaEditor from "@/components/SchemaEditor";
import ResultsTable from "@/components/ResultsTable";
import { SchemaType, DiffResult } from "@/lib/types";
import { SAMPLES } from "@/lib/samples";

const SCHEMA_TYPES: { id: SchemaType; label: string }[] = [
  { id: "openapi",  label: "OpenAPI" },
  { id: "graphql",  label: "GraphQL" },
  { id: "grpc",     label: "gRPC / Protobuf" },
];

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function PlaygroundPage() {
  const [schemaType, setSchemaType] = useState<SchemaType>("openapi");
  const [base, setBase] = useState(SAMPLES.openapi.base);
  const [head, setHead] = useState(SAMPLES.openapi.head);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<DiffResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  function switchType(type: SchemaType) {
    setSchemaType(type);
    setBase(SAMPLES[type].base);
    setHead(SAMPLES[type].head);
    setResult(null);
    setError(null);
  }

  async function handleCompare() {
    setLoading(true);
    setResult(null);
    setError(null);

    try {
      const res = await fetch(`${API_URL}/api/compare`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ schema_type: schemaType, base_content: base, head_content: head }),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error ?? "Unknown error");
        return;
      }

      setResult(data as DiffResult);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Network error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      {/* Header */}
      <header className="border-b border-slate-800 bg-slate-900">
        <div className="max-w-[1400px] mx-auto px-6 py-4 flex items-center justify-between flex-wrap gap-4">
          <div className="flex items-center gap-3">
            <span className="text-xl font-bold">
              <span className="text-indigo-400">drift-guard</span> playground
            </span>
            <span className="text-[11px] font-medium bg-slate-700 text-slate-400 px-2 py-0.5 rounded-full">
              beta
            </span>
          </div>

          <div className="flex gap-1 bg-slate-950 rounded-lg p-1">
            {SCHEMA_TYPES.map((t) => (
              <button
                key={t.id}
                onClick={() => switchType(t.id)}
                className={`px-4 py-1.5 rounded-md text-sm font-medium transition-all ${
                  schemaType === t.id
                    ? "bg-indigo-600 text-white"
                    : "text-slate-400 hover:text-slate-200 hover:bg-slate-800"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
        </div>
      </header>

      <main className="max-w-[1400px] mx-auto px-6 py-6 space-y-5">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
          <SchemaEditor
            label="Base Schema"
            subtitle="— current / stable version"
            value={base}
            onChange={setBase}
            schemaType={schemaType}
          />
          <SchemaEditor
            label="Head Schema"
            subtitle="— new / proposed version"
            value={head}
            onChange={setHead}
            schemaType={schemaType}
          />
        </div>

        <div className="flex justify-center">
          <button
            onClick={handleCompare}
            disabled={loading}
            className="px-10 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-60 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-colors text-base"
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
                </svg>
                Comparing…
              </span>
            ) : "Compare"}
          </button>
        </div>

        {error && (
          <div className="rounded-lg border border-red-700 bg-red-950 px-4 py-3 text-red-400 text-sm font-mono whitespace-pre-wrap">
            {error}
          </div>
        )}

        {result && <ResultsTable result={result} />}
      </main>
    </div>
  );
}
