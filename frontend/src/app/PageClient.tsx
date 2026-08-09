'use client';

import { useState, useCallback } from "react";
import Navbar from "./Navbar";
import ProductList from "./ProductList";
import CartDrawer from "./CartDrawer";

interface Cart {
    id: string;
    items: any[];
    total_price: number;
    status: string;
}

export default function PageClient({
    initialProducts,
    userId,
}: {
    initialProducts: any[];
    userId: string;
}) {
    const [cart, setCart] = useState<Cart | null>(null);
    const [cartOpen, setCartOpen] = useState(false);

    const handleCartUpdate = useCallback((updatedCart: Cart | null) => {
        setCart(updatedCart);
    }, []);

    const handleAddToCart = useCallback((updatedCart: Cart) => {
        setCart(updatedCart);
        // Briefly flash the cart icon open so users see the update
        setCartOpen(true);
    }, []);

    const itemCount = cart?.items?.reduce((sum, item) => sum + item.quantity, 0) || 0;

    return (
        <main className="min-h-screen">
            <Navbar
                cartCount={itemCount}
                onCartClick={() => setCartOpen(true)}
            />

            <div className="max-w-7xl mx-auto px-6 pb-20">
                <header className="mb-12 text-center">
                    <h2 className="text-5xl font-extrabold mb-4 text-gradient">
                        Discover Premium Tech
                    </h2>
                    <p className="text-gray-400 max-w-2xl mx-auto text-lg">
                        Experience our microservice architecture in action. Every purchase reserves real inventory in the background via gRPC.
                    </p>
                </header>

                {initialProducts.length === 0 ? (
                    <div className="text-center p-12 glass rounded-3xl">
                        <h3 className="text-2xl font-bold mb-2">No products found</h3>
                        <p className="text-gray-400">Run the seeder script to populate the database.</p>
                    </div>
                ) : (
                    <ProductList
                        initialProducts={initialProducts}
                        userId={userId}
                        onCartUpdate={handleAddToCart}
                    />
                )}
            </div>

            <CartDrawer
                isOpen={cartOpen}
                onClose={() => setCartOpen(false)}
                cart={cart}
                userId={userId}
                onCartUpdate={handleCartUpdate}
            />
        </main>
    );
}
