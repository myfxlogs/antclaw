package com.antclaw.alfq.data.repository

import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostVisibility
import org.junit.Assert.*
import org.junit.Test

class SocialRepositoryMappingTest {

    @Test fun `mapPostType signal_card → SIGNAL_CARD`() =
        assertEquals(PostType.SIGNAL_CARD, mapPostType("signal_card"))

    @Test fun `mapPostType chart_share → CHART_SHARE`() =
        assertEquals(PostType.CHART_SHARE, mapPostType("chart_share"))

    @Test fun `mapPostType share → SHARE`() =
        assertEquals(PostType.SHARE, mapPostType("share"))

    @Test fun `mapPostType unknown → TEXT`() =
        assertEquals(PostType.TEXT, mapPostType(""))

    @Test fun `mapVisibility public → PUBLIC`() =
        assertEquals(PostVisibility.PUBLIC, mapVisibility("public"))

    @Test fun `mapVisibility followers → FOLLOWERS_ONLY`() =
        assertEquals(PostVisibility.FOLLOWERS_ONLY, mapVisibility("followers"))

    @Test fun `mapVisibility circle → CIRCLE_ONLY`() =
        assertEquals(PostVisibility.CIRCLE_ONLY, mapVisibility("circle"))

    @Test fun `mapVisibility unknown → PUBLIC`() =
        assertEquals(PostVisibility.PUBLIC, mapVisibility("unknown"))
}
