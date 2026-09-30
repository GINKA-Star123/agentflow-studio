"use client";

import { useEffect, useRef } from "react";

import { useAuthStore } from "@/stores/auth-store";

type ProvidersProps = {
  children: React.ReactNode;
};

export function Providers({children}:ProvidersProps){
    const initializedRef = useRef(false);
    const initialize = useAuthStore((state) => state.initialize);

    useEffect(() => {
        if (initializedRef.current){
            return ;
        }

        initializedRef.current = true;
        void initialize();
    },[initialize]);

    return children;
}