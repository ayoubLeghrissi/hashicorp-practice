'use client';

import { motion } from "framer-motion";
import { ShoppingBag, Check } from "lucide-react";
import { useState } from "react";
import { addToCart } from "./actions";

export default function ProductList({
    initialProducts,
    userId,
    onCartUpdate,
}: {
    initialProducts: any[];
    userId: string;
    onCartUpdate?: (cart: any) => void;
}) {
    const [addingToCart, setAddingToCart] = useState<Record<string, boolean>>({});
    const [added, setAdded] = useState<Record<string, boolean>>({});

    const handleAddToCart = async (productId: string) => {
        setAddingToCart(prev => ({ ...prev, [productId]: true }));
        try {
            const updatedCart = await addToCart(userId, productId, 1);
            setAdded(prev => ({ ...prev, [productId]: true }));
            setTimeout(() => setAdded(prev => ({ ...prev, [productId]: false })), 2000);
            // Bubble the updated cart up so parent can update count + drawer
            onCartUpdate?.(updatedCart);
        } catch (error: any) {
            alert(error.message || "Failed to add to cart");
        } finally {
            setAddingToCart(prev => ({ ...prev, [productId]: false }));
        }
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {initialProducts.map((product, index) => (
                <motion.div
                    key={product.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.1 }}
                    whileHover={{ y: -5 }}
                    className="glass rounded-3xl overflow-hidden flex flex-col group relative"
                >
                    {/* Card Image */}
                    <div className="h-48 w-full bg-white/5 relative overflow-hidden flex items-center justify-center p-6">
                        <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                        <img
                            src={product.image_url || `https://source.unsplash.com/random/400x300?${encodeURIComponent(product.category || 'tech')}`}
                            alt={product.name}
                            className="w-full h-full object-cover rounded-xl shadow-lg transform group-hover:scale-105 transition-transform duration-500"
                            onError={(e) => { e.currentTarget.style.display = 'none'; }}
                        />
                    </div>

                    {/* Details */}
                    <div className="p-6 flex-1 flex flex-col">
                        <div className="text-xs font-semibold tracking-wider text-blue-400 mb-2 uppercase">
                            {product.category || "Uncategorized"}
                        </div>
                        <h3 className="text-xl font-bold mb-2 line-clamp-1">{product.name}</h3>
                        <p className="text-gray-400 text-sm mb-4 line-clamp-2 flex-1">
                            {product.description}
                        </p>
                        <div className="flex items-center justify-between mt-auto pt-4 border-t border-white/10">
                            <span className="text-2xl font-black">${parseFloat(product.price).toFixed(2)}</span>
                            <button
                                disabled={addingToCart[product.id] || added[product.id]}
                                onClick={() => handleAddToCart(product.id)}
                                className={`p-3 rounded-2xl flex items-center justify-center transition-all duration-300 ${added[product.id]
                                    ? 'bg-green-500 text-white'
                                    : 'bg-white/10 hover:bg-white/20 text-white'
                                    }`}
                            >
                                {added[product.id] ? (
                                    <Check className="w-5 h-5" />
                                ) : addingToCart[product.id] ? (
                                    <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                ) : (
                                    <ShoppingBag className="w-5 h-5 group-hover:scale-110 transition-transform" />
                                )}
                            </button>
                        </div>
                    </div>
                </motion.div>
            ))}
        </div>
    );
}
