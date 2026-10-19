"use client";
import { cn } from "@/lib/utils";
import { ReactNode } from "react";

export function Card({ children, className, onClick }: { children: ReactNode; className?: string; onClick?: () => void }) {
  return (
    <div onClick={onClick}
      className={cn("bg-white border border-zu-border rounded-lg shadow-zu-sm",
        onClick && "cursor-pointer transition-all hover:shadow-zu-md", className)}>
      {children}
    </div>
  );
}

type BadgeVariant = "blue" | "green" | "amber" | "red" | "gray" | "teal";
const BadgeStyles: Record<BadgeVariant, string> = {
  blue: "bg-[#EFF6FF] text-[#1D4ED8] border-[#BFDBFE]",
  green: "bg-[#F0FDF4] text-[#15803D] border-[#BBF7D0]",
  amber: "bg-[#FFFBEB] text-[#B45309] border-[#FDE68A]",
  red: "bg-[#FEF2F2] text-[#B91C1C] border-[#FECACA]",
  gray: "bg-[#F9FAFB] text-[#374151] border-[#E5E7EB]",
  teal: "bg-[#F0FDFA] text-[#0F766E] border-[#99F6E4]",
};

export function Badge({ children, variant = "gray", dot }: { children: ReactNode; variant?: BadgeVariant; dot?: boolean }) {
  return (
    <span className={cn("inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold border", BadgeStyles[variant])}>
      {dot && <span className="w-1.5 h-1.5 rounded-full" style={{ background: "currentColor" }} />}
      {children}
    </span>
  );
}

export function ScoreBar({ value, size = "md" }: { value: number; size?: "sm" | "md" }) {
  const color = value >= 85 ? "#16A34A" : value >= 65 ? "#D97706" : "#DC2626";
  const h = size === "sm" ? "h-[3px]" : "h-[5px]";
  return (
    <div className="flex items-center gap-2">
      <div className={cn("flex-1 bg-zu-surface2 rounded-full overflow-hidden", h)}>
        <div className="h-full rounded-full transition-all duration-500" style={{ width: `${value}%`, backgroundColor: color }} />
      </div>
      <span className="text-[12px] font-semibold font-mono min-w-[28px] text-right" style={{ color }}>{value}</span>
    </div>
  );
}

export function KPICard({ label, value, sub, trend, icon: Icon, valueColor }: { label: string; value: string | number; sub?: string; trend?: "up" | "down" | "flat"; icon?: any; valueColor?: string }) {
  return (
    <Card className="p-4">
      <div className="flex justify-between items-start mb-2">
        <span className="text-[12px] text-zu-text2 font-medium">{label}</span>
        {Icon && <Icon size={13} className="text-zu-text3" />}
      </div>
      <div className="text-[26px] font-bold tracking-tight font-mono leading-none mb-1.5" style={{ color: valueColor || "var(--text-1)" }}>{value}</div>
      {sub && (
        <div className={cn("text-[11px] flex items-center gap-1", { "text-zu-green": trend === "up", "text-zu-red": trend === "down", "text-zu-text3": !trend || trend === "flat" })}>
          {trend === "up" && "↑"}{trend === "down" && "↓"} {sub}
        </div>
      )}
    </Card>
  );
}

type BtnVariant = "primary" | "secondary" | "ghost" | "danger";
export function Button({ children, variant = "primary", size = "md", onClick, disabled, className, type = "button" }: { children: ReactNode; variant?: BtnVariant; size?: "sm" | "md"; onClick?: () => void; disabled?: boolean; className?: string; type?: "button" | "submit" }) {
  const variants: Record<BtnVariant, string> = {
    primary: "bg-zu-accent text-white hover:bg-blue-600 border-transparent",
    secondary: "bg-white text-zu-text1 hover:bg-zu-surface2 border-zu-border",
    ghost: "bg-transparent text-zu-text2 hover:bg-zu-surface2 border-transparent",
    danger: "bg-zu-red text-white hover:bg-red-600 border-transparent",
  };
  const sizes = { sm: "px-3 py-1.5 text-[11px]", md: "px-4 py-1.5 text-[12px]" };
  return (
    <button type={type} onClick={onClick} disabled={disabled}
      className={cn("inline-flex items-center gap-1.5 font-semibold rounded-md border transition-colors disabled:opacity-50 disabled:cursor-not-allowed", variants[variant], sizes[size], className)}>
      {children}
    </button>
  );
}

export function Table({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn("border border-zu-border rounded-lg overflow-hidden shadow-zu-sm bg-white", className)}>
      <table className="w-full border-collapse">{children}</table>
    </div>
  );
}
export function Th({ children }: { children: ReactNode }) {
  return <th className="px-3 py-2 text-left text-[11px] font-semibold text-zu-text2 tracking-wide bg-zu-surface2 border-b border-zu-border whitespace-nowrap">{children}</th>;
}
export function Td({ children, className }: { children: ReactNode; className?: string }) {
  return <td className={cn("px-3 py-2.5 text-[12px] border-b border-zu-border last:border-0", className)}>{children}</td>;
}
export function Tr({ children, onClick }: { children: ReactNode; onClick?: () => void }) {
  return <tr onClick={onClick} className={cn("transition-colors", onClick && "cursor-pointer hover:bg-zu-surface2")}>{children}</tr>;
}

export function Empty({ icon: Icon, title, description, action }: { icon: any; title: string; description?: string; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3 text-center">
      <Icon size={36} className="text-zu-border" />
      <div>
        <div className="text-[14px] font-semibold text-zu-text2">{title}</div>
        {description && <div className="text-[12px] text-zu-text3 mt-0.5 max-w-xs">{description}</div>}
      </div>
      {action}
    </div>
  );
}

export function Spinner({ size = 16 }: { size?: number }) {
  return (
    <div className="flex items-center justify-center p-4">
      <div className="rounded-full border-2 border-zu-accent/20 border-t-zu-accent animate-spin" style={{ width: size, height: size }} />
    </div>
  );
}

export function PageHeader({ title, description, action }: { title: string; description?: string; action?: ReactNode }) {
  return (
    <div className="flex justify-between items-start mb-5">
      <div>
        <h1 className="text-[17px] font-bold text-zu-text1 tracking-tight">{title}</h1>
        {description && <p className="text-[12px] text-zu-text2 mt-0.5">{description}</p>}
      </div>
      {action}
    </div>
  );
}
