import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import SchemaEditor from "@/components/SchemaEditor";

// next/dynamic loads lazily — override to return a synchronous textarea mock
vi.mock("next/dynamic", () => ({
  default: () =>
    ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
      <textarea
        data-testid="monaco-editor"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    ),
}));

describe("SchemaEditor", () => {
  it("renders label text", () => {
    render(
      <SchemaEditor
        label="Base Schema"
        subtitle="— current version"
        value="openapi: 3.0.0"
        onChange={() => {}}
        schemaType="openapi"
      />
    );
    expect(screen.getByText("Base Schema")).toBeInTheDocument();
  });

  it("renders subtitle text", () => {
    render(
      <SchemaEditor
        label="Head Schema"
        subtitle="— new version"
        value=""
        onChange={() => {}}
        schemaType="graphql"
      />
    );
    expect(screen.getByText("— new version")).toBeInTheDocument();
  });

  it("passes value to the editor", () => {
    render(
      <SchemaEditor
        label="Base Schema"
        subtitle=""
        value="type Query { ping: String }"
        onChange={() => {}}
        schemaType="graphql"
      />
    );
    const editor = screen.getByTestId("monaco-editor") as HTMLTextAreaElement;
    expect(editor.value).toBe("type Query { ping: String }");
  });

  it("calls onChange when editor content changes", () => {
    const handleChange = vi.fn();
    render(
      <SchemaEditor
        label="Base Schema"
        subtitle=""
        value="initial"
        onChange={handleChange}
        schemaType="openapi"
      />
    );
    const editor = screen.getByTestId("monaco-editor");
    fireEvent.change(editor, { target: { value: "updated" } });
    expect(handleChange).toHaveBeenCalledWith("updated");
  });
});
