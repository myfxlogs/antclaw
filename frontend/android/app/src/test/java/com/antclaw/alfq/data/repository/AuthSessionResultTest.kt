package com.antclaw.alfq.data.repository

import org.junit.Assert.*
import org.junit.Test

class AuthSessionResultTest {

    @Test fun `full construction with all fields`() {
        val result = AuthSessionResult(
            userId = "u123",
            accessToken = "acc.token.abc",
            refreshToken = "ref.token.def",
            displayName = "TraderJoe",
            codeId = "12345",
        )
        assertEquals("u123", result.userId)
        assertEquals("acc.token.abc", result.accessToken)
        assertEquals("ref.token.def", result.refreshToken)
        assertEquals("TraderJoe", result.displayName)
        assertEquals("12345", result.codeId)
    }

    @Test fun `default displayName and codeId are empty`() {
        val result = AuthSessionResult(
            userId = "u1",
            accessToken = "tok",
            refreshToken = "ref",
        )
        assertEquals("", result.displayName)
        assertEquals("", result.codeId)
    }

    @Test fun `equality works correctly`() {
        val a = AuthSessionResult("u1", "at", "rt")
        val b = AuthSessionResult("u1", "at", "rt")
        assertEquals(a, b)
        assertEquals(a.hashCode(), b.hashCode())
    }
}
